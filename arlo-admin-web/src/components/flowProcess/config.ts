import { randomLenNum } from '@/utils/random'

const getKey = () => {
  return `flk${randomLenNum(14)}`
}
export default {
  nodeKey: () => {
    return getKey()
  },
  nodes: [
    {
      name: '审批节点',
      defaultName: '审核人',
      placeholder: "请选择",
      type: 1,
      color: '#e6a23c',
      icon: 'Avatar',
      config: {
        // 审批人员类型 1指定成员，2主管，3角色，4发起人自选，5发起人自己，6连续多级主管
        setType: 2,
        // 成员列表：人员，角色
        nodeAssigneeList: [],
        // 指定主管：发起人的第N级主管
        examineLevel: 1,
        // 全员参与审批 1是，0否
        groupStrategy: 1,
        // 发起人自选：1自选一个人，2自选多个人，3自选角色
        selectMode: 1,
        // 发起人自选内容列表
        nodeCandidate: {
          // 类型：0人员，1角色
          type: 1,
          // 角色列表
          assignees: [
            // {
            //   id: "0",
            //   name: "系统管理员"
            // }
          ]
        },
        // 连续主管审批终点：0直到最上层主管，1自定义审批终点
        directorMode: 0,
        // 超时自动审批
        termAuto: true,
        // 审批期限（为 0 则不生效）
        term: 1,
        // 审批期限超时后执行：0自动通过，1自动拒绝
        termMode: 0,
        // 审批提醒
        remind: false,
        // 延时时间：1固定时长，2自动计算
        delayType: 1,
        // 多人审批时审批方式：1按顺序依次审批，2会签 (可同时审批，每个人必须审批通过)，3或签 (有一人审批通过即可)
        examineMode: 1,
        // 审批人与提交人为同一人时：0由发起人对自己审批，1自动跳过，2转交给直接上级审批，3转交给部门负责人审批
        approveSelf: 0,
        // 允许转交
        allowTransfer: true,
        // 允许加签/减签
        allowAppendNode: true,
        // 允许回退
        allowRollback: true,
        // 允许抄送
        allowCc: true,
        // 驳回策略：1驳回到发起人，2驳回到上一节点，3驳回到指定节点，4终止流程，5驳回到模型父节点
        rejectStrategy: 3,
        // 驳回重新审批策略：1继续往下执行，2回到原驳回节点
        rejectStart: 1,
        // 子表单  "1797256620887674882:测试设计表单"
        actionUrl: "",
        // 扩展配置
        extendConfig: {
          // 表单配置
          formConfig: [
            // {
            //   "label": "单行文本",
            //   "id": "1756290222360",
            //   "opera": "1"
            // }
          ],
          // 延时时间
          remindTime: ""
        },

        directorLevel: 1,
      }
    },
    {
      name: '抄送节点',
      defaultName: '抄送人',
      placeholder: "请选择",
      type: 2,
      color: '#409eff',
      icon: 'Promotion',
      config: {
        // 允许发起人自选抄送人
        allowSelection: true,
        // 审批提醒
        remind: false,
        // 抄送人员列表
        nodeAssigneeList: [],
      }
    },
    {
      name: '条件分支',
      defaultName: '条件路由',
      placeholder: "请设置条件",
      type: 4,
      color: '#67c23a',
      icon: 'Share',
      config: {
				conditionNodes: [
				  {
            nodeName: "条件1",
            nodeKey: getKey(),
            type: 3,
            // 优先级
            priorityLevel: 1,
            // 条件列表
            conditionList: []
          }, 
          {
            nodeName: "默认条件",
            nodeKey: `${getKey()}default`,
            type: 3,
            priorityLevel: 2,
            conditionList: []
          }
        ],
      }
    },
    {
      name: '子流程',
      defaultName: '子流程',
      placeholder: "请选择子流程规则",
      type: 5,
      color: 'rgb(146, 96, 250)',
      icon: 'Money',
      config: {
        // 子流程ID
        subProcessValue: "",
        // 子流程信息
        callProcess: ""
      }
    },
    {
      name: '延迟等待',
      defaultName: '延时处理',
      placeholder: "请设置延迟时间",
      type: 6,
      color: '#f56c6c',
      icon: 'Clock',
      config: {
        delayType: 1,
        extendConfig: {
          time: "1:m"
        },
      }
    },
    {
      name: '触发器',
      defaultName: '触发器',
      placeholder: '请选择触发器规则',
      type: 7,
      color: 'rgb(43, 181, 139)',
      icon: 'SetUp',
      config:       {
        // 执行方式 1立即执行，2延迟执行
        triggerType: 1,
        // 延迟方式 1固定时长（与延时节点一致）
        delayType: 1,
        // 扩展
        extendConfig: {
          // 延长时间，如 1:m
          time: '1:m',
          // JSON 附加参数（字符串）
          args: '',
          // HTTP(S) URL
          trigger: '',
          // GET / POST / PUT
          method: 'POST',
        },
      }
    },
    {
      name: '并行分支',
      defaultName: '并行路由',
      placeholder: "并行任务（同时进行）",
      type: 8,
      color: 'rgb(98, 106, 239)',
      icon: 'Operation',
      config: {
        parallelNodes: [
          {
            nodeName: "并行分支1",
            nodeKey: getKey(),
            type: 3,
            priorityLevel: 1,
            conditionList: [],
          }, 
          {
            nodeName: "并行分支2",
            nodeKey: getKey(),
            type: 3,
            priorityLevel: 2,
            conditionList: []
          }
        ],
      }
    },
    {
      name: '包容分支',
      defaultName: '包容路由',
      placeholder: "请设置条件",
      type: 9,
      color: 'rgb(52, 93, 162)',
      icon: 'CopyDocument',
      config: {
        inclusiveNodes: [
          {
            nodeName: "包容条件1",
            nodeKey: getKey(),
            type: 3,
            priorityLevel: 1,
            conditionList: []
          }, 
          {
            nodeName: "默认条件",
            nodeKey: `${getKey()}default`,
            type: 3,
            priorityLevel: 2,
            conditionList: []
          }
        ],
      }
    },
    {
      name: '路由分支',
      defaultName: '路由分支',
      placeholder: "请设置路由节点",
      type: 23,
      color: 'rgb(220, 38, 38)',
      icon: 'Guide',
      config: {
        routeNodes: [],
      }
    },
    {
      name: '自动通过',
      defaultName: '自动通过',
      placeholder: "自动通过",
      type: 30,
      color: '#67c23a',
      icon: 'CircleCheckFilled',
      config: {}
    },
    {
      name: '自动拒绝',
      defaultName: '自动拒绝',
      placeholder: "自动拒绝",
      type: 31,
      color: '#f56c6c',
      icon: 'CircleCloseFilled',
      config: {}
    },
  ],
}