package examples

import (
	"fmt"
	"testing"

	"github.com/rs/zerolog"
	"github.com/sunvim/utils/logger"
)

func TestLoggerExample(t *testing.T) {
	t.Log("Testing logger package")

	// 测试默认配置
	defaultCfg := logger.DefaultLoggingConfig()
	if defaultCfg.Level != "info" {
		t.Fatalf("Expected default level 'info', got '%s'", defaultCfg.Level)
	}

	// 创建日志管理器
	cfg := &logger.LoggingConfig{
		Level:  "debug",
		Format: "text",
		Output: "stdout",
	}

	manager, err := logger.NewLoggerManager(cfg)
	if err != nil {
		t.Fatalf("Failed to create logger manager: %v", err)
	}
	defer manager.Close()

	log := manager.GetLogger()

	// 测试各种日志级别
	log.Debug("这是一条调试消息")
	log.Info("这是一条信息消息")
	log.Warn("这是一条警告消息")
	log.Error("这是一条错误消息")

	t.Log("基本日志记录测试通过")

	// 测试格式化日志
	log.Debugf("调试消息: %s = %d", "value", 42)
	log.Infof("信息消息: %s", "formatted")
	log.Warnf("警告消息: %v", map[string]int{"key": 123})
	log.Errorf("错误消息: %s", "something went wrong")

	t.Log("格式化日志测试通过")

	// 测试带字段的日志
	logWithField := log.WithField("component", "test")
	logWithField.Info("带字段的日志消息")

	logWithFields := log.WithFields(map[string]interface{}{
		"user_id":    123,
		"session_id": "abc123",
		"action":     "login",
	})
	logWithFields.Info("带多个字段的日志消息")

	t.Log("带字段的日志测试通过")

	// 测试带错误的日志
	testErr := fmt.Errorf("测试错误")
	logWithError := log.WithError(testErr)
	logWithError.Error("发生了错误")

	t.Log("带错误的日志测试通过")
}

func TestLoggerGlobalFunctions(t *testing.T) {
	t.Log("Testing global logger functions")

	// 初始化全局日志器
	cfg := &logger.LoggingConfig{
		Level:  "info",
		Format: "console",
		Output: "stderr",
	}

	err := logger.InitLogger(cfg)
	if err != nil {
		t.Fatalf("Failed to init global logger: %v", err)
	}
	defer logger.CloseGlobalLogger()

	// 获取全局日志器
	globalLog := logger.GetGlobalLogger()
	globalLog.Info("全局日志器测试消息")

	// 获取带组件的日志器
	componentLog := logger.GetLoggerWithComponent("test-component")
	componentLog.Info("带组件的日志消息")

	// 获取带模块的日志器
	moduleLog := logger.GetLoggerWithModule("test-module")
	moduleLog.Info("带模块的日志消息")

	t.Log("全局日志器函数测试通过")

	// 测试日志级别设置
	err = logger.SetGlobalLogLevel("debug")
	if err != nil {
		t.Fatalf("Failed to set global log level: %v", err)
	}

	globalLog.Debug("这条调试消息现在应该显示")
	t.Log("日志级别设置测试通过")
}

func TestLoggerMetricsHook(t *testing.T) {
	t.Log("Testing metrics hook")

	// 创建度量钩子
	metricsHook := logger.NewMetricsHook()

	// 模拟日志事件
	metricsHook.Fire(zerolog.InfoLevel, "info message")
	metricsHook.Fire(zerolog.WarnLevel, "warn message")
	metricsHook.Fire(zerolog.ErrorLevel, "error message")
	metricsHook.Fire(zerolog.DebugLevel, "debug message")
	metricsHook.Fire(zerolog.InfoLevel, "another info message")

	// 检查计数
	errors, warns, infos, debugs := metricsHook.GetCounts()
	if errors != 1 {
		t.Fatalf("Expected 1 error, got %d", errors)
	}
	if warns != 1 {
		t.Fatalf("Expected 1 warning, got %d", warns)
	}
	if infos != 2 {
		t.Fatalf("Expected 2 info messages, got %d", infos)
	}
	if debugs != 1 {
		t.Fatalf("Expected 1 debug message, got %d", debugs)
	}

	t.Log("度量钩子测试通过")
}

func TestLoggerRequestIDHook(t *testing.T) {
	t.Log("Testing request ID hook")

	// 创建请求ID钩子
	requestIDHook := logger.NewRequestIDHook()

	// 测试钩子执行
	err := requestIDHook.Fire(zerolog.InfoLevel, "test message")
	if err != nil {
		t.Fatalf("RequestID hook should not return error: %v", err)
	}

	t.Log("请求ID钩子测试通过")
}

func TestLoggerConfiguration(t *testing.T) {
	t.Log("Testing logger configuration")

	// 测试JSON格式
	jsonCfg := &logger.LoggingConfig{
		Level:  "warn",
		Format: "json",
		Output: "stdout",
	}

	jsonManager, err := logger.NewLoggerManager(jsonCfg)
	if err != nil {
		t.Fatalf("Failed to create JSON logger: %v", err)
	}
	defer jsonManager.Close()

	jsonLog := jsonManager.GetLogger()
	jsonLog.Warn("JSON格式的警告消息")

	t.Log("JSON格式配置测试通过")

	// 测试不同日志级别
	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		cfg := &logger.LoggingConfig{
			Level:  level,
			Format: "console",
			Output: "stderr",
		}

		mgr, err := logger.NewLoggerManager(cfg)
		if err != nil {
			t.Fatalf("Failed to create logger with level %s: %v", level, err)
		}

		log := mgr.GetLogger()
		log.Info(fmt.Sprintf("测试日志级别: %s", level))
		mgr.Close()
	}

	t.Log("不同日志级别配置测试通过")
}
