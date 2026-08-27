#!/bin/bash
# ============================================================
# arlo-admin API 全端点测试脚本
# 用法: bash scripts/test_api.sh [BASE_URL]
# 默认 BASE_URL=http://localhost:8090/api/v1
# ============================================================
set -e

BASE_URL="${1:-http://localhost:8090/api/v1}"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

PASS=0
FAIL=0
TOKEN=""

# 颜色输出
ok()   { echo -e "${GREEN}[PASS]${NC} $1"; }
fail() { echo -e "${RED}[FAIL]${NC} $1"; ((FAIL++)) && return 1; }
info() { echo -e "${CYAN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
section() { echo -e "\n${CYAN}═══════════════════════════════════${NC}"; echo -e "${CYAN}  $1${NC}"; echo -e "${CYAN}═══════════════════════════════════${NC}"; }

# 发送请求并校验
# check POST|GET|PUT|DELETE path [json_body] [expected_key]
check() {
    local method="$1"
    local path="$2"
    local body="${3:-}"
    local expected_key="${4:-code}"

    local curl_cmd="curl -s -w '%{http_code}'"
    if [ -n "$TOKEN" ]; then
        curl_cmd="$curl_cmd -H 'Authorization: Bearer $TOKEN'"
    fi
    curl_cmd="$curl_cmd -H 'Content-Type: application/json'"

    case "$method" in
        POST)   curl_cmd="$curl_cmd -X POST -d '$body'" ;;
        PUT)    curl_cmd="$curl_cmd -X PUT -d '$body'" ;;
        DELETE) curl_cmd="$curl_cmd -X DELETE" ;;
    esac

    local full_url="${BASE_URL}${path}"
    local raw_output
    raw_output=$(eval "$curl_cmd '$full_url'" 2>/dev/null) || true

    local http_code="${raw_output: -3}"
    local json_body="${raw_output:0:-3}"

    if [ -z "$json_body" ]; then
        fail "$method $path — 空响应 (HTTP $http_code)"
        return 1
    fi

    # 提取 code 字段
    local resp_code
    resp_code=$(echo "$json_body" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code',''))" 2>/dev/null || echo "")

    if [ "$resp_code" = "200" ] || [ "$resp_code" = "200 " ]; then
        local msg
        msg=$(echo "$json_body" | python3 -c "import sys,json; print(json.load(sys.stdin).get('msg',''))" 2>/dev/null || echo "")
        ok "$method $path → $msg"
        ((PASS++))
        echo "$json_body"
        return 0
    else
        fail "$method $path — code=$resp_code (HTTP $http_code)"
        echo "$json_body"
        ((FAIL++))
        return 1
    fi
}

# 带预期校验
check_custom() {
    local method="$1" path="$2" body="${3:-}" expected="${4:-200}" desc="${5:-}"
    local full_url="${BASE_URL}${path}"

    local curl_cmd="curl -s"
    if [ -n "$TOKEN" ]; then
        curl_cmd="$curl_cmd -H 'Authorization: Bearer $TOKEN'"
    fi
    curl_cmd="$curl_cmd -H 'Content-Type: application/json'"

    case "$method" in
        POST)   curl_cmd="$curl_cmd -X POST -d '$body'" ;;
        PUT)    curl_cmd="$curl_cmd -X PUT -d '$body'" ;;
        DELETE) curl_cmd="$curl_cmd -X DELETE" ;;
    esac

    local json_body
    json_body=$(eval "$curl_cmd '$full_url'" 2>/dev/null) || true

    if [ -z "$json_body" ]; then
        fail "$method $path — 空响应 ($desc)"
        return 1
    fi

    local resp_code
    resp_code=$(echo "$json_body" | python3 -c "import sys,json; print(json.load(sys.stdin).get('code',''))" 2>/dev/null || echo "")

    if [ "$resp_code" = "$expected" ]; then
        ok "$method $path ($desc)"
        ((PASS++))
        return 0
    else
        fail "$method $path ($desc) — expected $expected got $resp_code"
        ((FAIL++))
        return 1
    fi
}

# ============================================================
# 1. 健康检查
# ============================================================
section "健康检查"
check GET "/../health" "" ""

# ============================================================
# 2. 认证模块 (auth) — 4 端点
# ============================================================
section "Auth 模块（4 端点）"

info "登录获取 token..."
LOGIN_RESP=$(check POST "/auth/login" '{"username":"admin","password":"admin123"}' "code")
TOKEN=$(echo "$LOGIN_RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('accessToken',''))" 2>/dev/null || echo "")

if [ -z "$TOKEN" ]; then
    warn "登录失败，后续需认证的接口将跳过"
else
    info "Token: ${TOKEN:0:20}..."
    check GET "/auth/info" "" "code"
    check POST "/auth/refresh" "{\"refreshToken\":\"$(echo "$LOGIN_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('refreshToken',''))" 2>/dev/null || echo "")\"}" "code"
    check POST "/auth/logout" "" "code"
fi

# ============================================================
# 3. 用户管理 (system/user) — 6 端点
# ============================================================
section "System/User 模块（6 端点）"

check GET "/system/user/list?page=1&pageSize=5" "" "code"
check GET "/system/user/1" "" "code"

CREATE_USER_RESP=$(check POST "/system/user" '{"username":"testuser","password":"123456","nickname":"测试用户","deptId":1,"status":1,"roleIds":[2],"postIds":[]}' "code")
NEW_USER_ID=$(echo "$CREATE_USER_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null || echo "")

if echo "$CREATE_USER_RESP" | python3 -c "import sys,json; j=json.load(sys.stdin); sys.exit(0 if j.get('code')==200 else 1)" 2>/dev/null; then
    NEW_USER_ID=$(echo "$CREATE_USER_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null || echo "")
    info "创建用户成功, ID=$NEW_USER_ID"

    if [ -n "$NEW_USER_ID" ] && [ "$NEW_USER_ID" != "None" ]; then
        check PUT "/system/user" "{\"id\":$NEW_USER_ID,\"nickname\":\"测试用户-已更新\",\"deptId\":1,\"status\":1,\"roleIds\":[2],\"postIds\":[]}" "code"
        check PUT "/system/user/password" "{\"id\":$NEW_USER_ID,\"password\":\"654321\"}" "code"
        check DELETE "/system/user/$NEW_USER_ID" "" "code"
    fi
fi

# ============================================================
# 4. 角色管理 (system/role) — 8 端点
# ============================================================
section "System/Role 模块（8 端点）"

check GET "/system/role/list?page=1&pageSize=10" "" "code"
check GET "/system/role/all" "" "code"
check GET "/system/role/1" "" "code"
check GET "/system/role/1/menus" "" "code"

CREATE_ROLE_RESP=$(check POST "/system/role" '{"roleName":"测试角色","roleCode":"test_role","sort":99,"status":1,"remark":"curl测试"}' "code")
NEW_ROLE_ID=$(echo "$CREATE_ROLE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null || echo "")

if [ -n "$NEW_ROLE_ID" ] && [ "$NEW_ROLE_ID" != "None" ]; then
    info "创建角色成功, ID=$NEW_ROLE_ID"
    check PUT "/system/role" "{\"id\":$NEW_ROLE_ID,\"roleName\":\"测试角色-已更新\",\"roleCode\":\"test_role\",\"sort\":88,\"status\":1}" "code"
    check POST "/system/role/assignMenus" "{\"roleId\":$NEW_ROLE_ID,\"menuIds\":[]}" "code"
    check DELETE "/system/role/$NEW_ROLE_ID" "" "code"
fi

# ============================================================
# 5. 部门管理 (system/dept) — 4 端点
# ============================================================
section "System/Dept 模块（4 端点）"

check GET "/system/dept/tree" "" "code"

CREATE_DEPT_RESP=$(check POST "/system/dept" '{"deptName":"测试部门","parentId":0,"sort":99,"leader":"","phone":"","email":"","status":1}' "code")
NEW_DEPT_ID=$(echo "$CREATE_DEPT_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null || echo "")

if [ -n "$NEW_DEPT_ID" ] && [ "$NEW_DEPT_ID" != "None" ]; then
    info "创建部门成功, ID=$NEW_DEPT_ID"
    check PUT "/system/dept" "{\"id\":$NEW_DEPT_ID,\"deptName\":\"测试部门-已更新\",\"parentId\":0,\"sort\":88,\"status\":1}" "code"
    check DELETE "/system/dept/$NEW_DEPT_ID" "" "code"
fi

# ============================================================
# 6. 菜单管理 (system/menu) — 4 端点
# ============================================================
section "System/Menu 模块（4 端点）"

check GET "/system/menu/tree" "" "code"

CREATE_MENU_RESP=$(check POST "/system/menu" '{"menuName":"测试菜单","parentId":0,"type":0,"path":"/test","sort":99,"icon":"","status":1,"visible":1,"keepAlive":0}' "code")
NEW_MENU_ID=$(echo "$CREATE_MENU_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null || echo "")

if [ -n "$NEW_MENU_ID" ] && [ "$NEW_MENU_ID" != "None" ]; then
    info "创建菜单成功, ID=$NEW_MENU_ID"
    check PUT "/system/menu" "{\"id\":$NEW_MENU_ID,\"menuName\":\"测试菜单-已更新\",\"parentId\":0,\"type\":0,\"path\":\"/test\",\"sort\":88,\"status\":1,\"visible\":1,\"keepAlive\":0}" "code"
    check DELETE "/system/menu/$NEW_MENU_ID" "" "code"
fi

# ============================================================
# 7. 岗位管理 (system/post) — 6 端点
# ============================================================
section "System/Post 模块（6 端点）"

check GET "/system/post/list?page=1&pageSize=10" "" "code"
check GET "/system/post/all" "" "code"
check GET "/system/post/1" "" "code"

CREATE_POST_RESP=$(check POST "/system/post" '{"postName":"测试岗位","postCode":"test_post","sort":99,"status":1}' "code")
NEW_POST_ID=$(echo "$CREATE_POST_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null || echo "")

if [ -n "$NEW_POST_ID" ] && [ "$NEW_POST_ID" != "None" ]; then
    info "创建岗位成功, ID=$NEW_POST_ID"
    check PUT "/system/post" "{\"id\":$NEW_POST_ID,\"postName\":\"测试岗位-已更新\",\"postCode\":\"test_post\",\"sort\":88,\"status\":1}" "code"
    check DELETE "/system/post/$NEW_POST_ID" "" "code"
fi

# ============================================================
# 8. 字典管理 (system/dict) — 10 端点
# ============================================================
section "System/Dict 模块（10 端点）"

# --- 字典类型 ---
check GET "/system/dict/type/list?page=1&pageSize=10" "" "code"
check GET "/system/dict/type/all" "" "code"

CREATE_TYPE_RESP=$(check POST "/system/dict/type" '{"dictName":"测试字典","dictType":"test_dict","status":1,"remark":"curl测试"}' "code")
NEW_TYPE_ID=$(echo "$CREATE_TYPE_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null || echo "")

if [ -n "$NEW_TYPE_ID" ] && [ "$NEW_TYPE_ID" != "None" ]; then
    info "创建字典类型成功, ID=$NEW_TYPE_ID"
    check PUT "/system/dict/type" "{\"id\":$NEW_TYPE_ID,\"dictName\":\"测试字典-已更新\",\"dictType\":\"test_dict\",\"status\":1}" "code"
fi

# --- 字典数据 ---
check GET "/system/dict/data/list?page=1&pageSize=10" "" "code"

if [ -n "$NEW_TYPE_ID" ] && [ "$NEW_TYPE_ID" != "None" ]; then
    check GET "/system/dict/data/type/$NEW_TYPE_ID" "" "code"

    CREATE_DATA_RESP=$(check POST "/system/dict/data" "{\"dictTypeId\":$NEW_TYPE_ID,\"dictLabel\":\"测试项\",\"dictValue\":\"test_value\",\"sort\":1,\"status\":1,\"isDefault\":0}" "code")
    NEW_DATA_ID=$(echo "$CREATE_DATA_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin).get('data',{}).get('id',''))" 2>/dev/null || echo "")

    if [ -n "$NEW_DATA_ID" ] && [ "$NEW_DATA_ID" != "None" ]; then
        info "创建字典数据成功, ID=$NEW_DATA_ID"
        check PUT "/system/dict/data" "{\"id\":$NEW_DATA_ID,\"dictTypeId\":$NEW_TYPE_ID,\"dictLabel\":\"测试项-已更新\",\"dictValue\":\"test_value\",\"sort\":2,\"status\":1}" "code"
        check DELETE "/system/dict/data/$NEW_DATA_ID" "" "code"
    fi

    # 清理字典类型
    check DELETE "/system/dict/type/$NEW_TYPE_ID" "" "code"
fi

# ============================================================
# 9. 未登录 → 401 校验
# ============================================================
section "401 认证校验"

SAVED_TOKEN="$TOKEN"
TOKEN=""
check_custom GET "/system/user/1" "" 1001 "未登录→应返回401"
TOKEN="$SAVED_TOKEN"

# ============================================================
# 10. 结果汇总
# ============================================================
section "测试结果汇总"
echo -e "  ${GREEN}通过: $PASS${NC}"
echo -e "  ${RED}失败: $FAIL${NC}"

TOTAL=$((PASS + FAIL))
echo -e "  总计: $TOTAL 个测试"

if [ "$FAIL" -eq 0 ]; then
    echo -e "\n${GREEN}  All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}  Some tests failed, check output above${NC}"
    exit 1
fi
