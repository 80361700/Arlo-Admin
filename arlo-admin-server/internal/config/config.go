package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Casbin   CasbinConfig   `mapstructure:"casbin"`
	Log      LogConfig      `mapstructure:"log"`
	Storage  StorageConfig  `mapstructure:"storage"`
	SMS      SMSConfig      `mapstructure:"sms"`
}

// SMSConfig 会员短信验证码
// provider: mock（仅打日志）| aliyun | tencent
type SMSConfig struct {
	Provider       string             `mapstructure:"provider"`
	CodeTTL        time.Duration      `mapstructure:"codeTTL"`        // 验证码有效期，默认 5m
	ResendInterval time.Duration      `mapstructure:"resendInterval"` // 重发间隔，默认 60s
	Aliyun         AliyunSMSConfig    `mapstructure:"aliyun"`
	Tencent        TencentSMSConfig   `mapstructure:"tencent"`
}

// AliyunSMSConfig 阿里云短信（dysmsapi）
type AliyunSMSConfig struct {
	AccessKeyID     string `mapstructure:"accessKeyId"`
	AccessKeySecret string `mapstructure:"accessKeySecret"`
	SignName        string `mapstructure:"signName"`
	TemplateCode    string `mapstructure:"templateCode"`
	// TemplateParamKey 模板变量名，默认 code（模板如 您的验证码为${code}）
	TemplateParamKey string `mapstructure:"templateParamKey"`
	RegionID         string `mapstructure:"regionId"` // 默认 cn-hangzhou
}

// TencentSMSConfig 腾讯云短信
type TencentSMSConfig struct {
	SecretID  string `mapstructure:"secretId"`
	SecretKey string `mapstructure:"secretKey"`
	AppID     string `mapstructure:"appId"`     // SmsSdkAppId
	SignName  string `mapstructure:"signName"`
	TemplateID string `mapstructure:"templateId"`
	Region    string `mapstructure:"region"` // 默认 ap-guangzhou
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
	EnableSwagger bool         `mapstructure:"enableSwagger"` // 是否开启 Swagger；空则 debug 开启、release 关闭
	CorsOrigins  []string      `mapstructure:"corsOrigins"`   // 允许的跨域来源；空则开发环境允许 *
}

type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	Charset         string        `mapstructure:"charset"`
	MaxIdleConns    int           `mapstructure:"maxIdleConns"`
	MaxOpenConns    int           `mapstructure:"maxOpenConns"`
	ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
	LogLevel        int           `mapstructure:"logLevel"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"poolSize"`
}

type JWTConfig struct {
	Secret        string        `mapstructure:"secret"`
	AccessExpire  time.Duration `mapstructure:"accessExpire"`
	RefreshExpire time.Duration `mapstructure:"refreshExpire"`
	Issuer        string        `mapstructure:"issuer"`
}

type CasbinConfig struct {
	ModelPath string `mapstructure:"modelPath"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	FilePath   string `mapstructure:"filePath"`
	MaxSize    int    `mapstructure:"maxSize"`
	MaxBackups int    `mapstructure:"maxBackups"`
	MaxAge     int    `mapstructure:"maxAge"`
	Compress   bool   `mapstructure:"compress"`
	Format     string `mapstructure:"format"`
}

// StorageConfig 文件存储配置
type StorageConfig struct {
	Driver      string             `mapstructure:"driver"`      // 存储驱动: local / oss
	MaxSize     int64              `mapstructure:"maxSize"`     // 最大上传大小(字节)
	AllowedExts []string           `mapstructure:"allowedExts"` // 允许的扩展名（不含点，小写）；空则用内置默认
	Local       LocalStorageConfig `mapstructure:"local"`       // 本地存储配置
	OSS         OSSStorageConfig   `mapstructure:"oss"`         // 阿里云OSS配置
}

// LocalStorageConfig 本地存储配置
type LocalStorageConfig struct {
	Path string `mapstructure:"path"` // 存储根目录
}

// OSSStorageConfig 阿里云OSS配置
type OSSStorageConfig struct {
	Endpoint        string `mapstructure:"endpoint"`        // OSS Endpoint，如 oss-cn-hangzhou.aliyuncs.com
	AccessKeyID     string `mapstructure:"accessKeyId"`     // AccessKey ID
	AccessKeySecret string `mapstructure:"accessKeySecret"` // AccessKey Secret
	BucketName      string `mapstructure:"bucketName"`      // Bucket名称
}

var GlobalConfig *Config

// Load 加载配置
// 1. 先加载 config.yaml（含所有默认值）
// 2. 如果指定了 prodPath，再合并 config.prod.yaml（仅覆盖不同的项）
func Load(basePath string, prodPath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(basePath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// 生产环境：合并 prod 配置（只覆盖 config.prod.yaml 中有的项，其余保留 config.yaml 默认值）
	if prodPath != "" {
		prodV := viper.New()
		prodV.SetConfigFile(prodPath)
		prodV.SetConfigType("yaml")
		if err := prodV.ReadInConfig(); err == nil {
			v.MergeConfigMap(prodV.AllSettings())
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	GlobalConfig = &cfg
	return &cfg, nil
}
