package config

import (
	"os"
	"testing"
)

func b2s(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func getTestConfig() *Config {
	wantOpts := newConfig()
	wantOpts.Name = "gleafd"
	wantOpts.Addr = "localhost:7890"
	wantOpts.Segment.Enable = false
	wantOpts.Segment.DBHost = "example.com:3306"
	wantOpts.Segment.DBName = "examaple"
	wantOpts.Segment.DBUser = "gleafd"
	wantOpts.Segment.DBPass = "12454656"
	wantOpts.Snowflake.Enable = false
	wantOpts.Snowflake.RedisAddresss = "localhost:8379"
	return wantOpts
}

func TestParseArgs(t *testing.T) {
	wantCfg := getTestConfig()
	args := []string{
		"--name=" + wantCfg.Name,
		"--addr=" + wantCfg.Addr,
		"--segment-enable=" + b2s(wantCfg.Segment.Enable),
		"--segment-db-host=" + wantCfg.Segment.DBHost,
		"--segment-db-user=" + wantCfg.Segment.DBUser,
		"--segment-db-pass=" + wantCfg.Segment.DBPass,
		"--segment-db-name=" + wantCfg.Segment.DBName,
		"--snowflake-enable=" + b2s(wantCfg.Snowflake.Enable),
		"--snowflake-redis-addr=" + wantCfg.Snowflake.RedisAddresss,
	}
	cfg, err := Load(args)
	if err != nil {
		t.Fatal(err)
	}
	if *cfg != *wantCfg {
		t.Fatalf("cfg = %v, want = %v", cfg, wantCfg)
	}
}

func TestParseArgsWithEnv(t *testing.T) {
	cliOpts := getTestConfig()
	args := []string{
		// "--name=" + cliOpts.Name,
		"--addr=" + cliOpts.Addr,
		"--segment-enable=" + b2s(cliOpts.Segment.Enable),
		"--segment-db-host=" + cliOpts.Segment.DBHost,
		"--segment-db-user=" + cliOpts.Segment.DBUser,
		"--segment-db-pass=" + cliOpts.Segment.DBPass,
		"--segment-db-name=" + cliOpts.Segment.DBName,
		// "--snowflake-enable=" + b2s(cliOpts.Snowflake.Enable),
		"--snowflake-redis-addr=" + cliOpts.Snowflake.RedisAddresss,
	}
	envOpts := newConfig()
	envOpts.Name = "myGleafd"
	envOpts.Addr = "fromenv:1239"
	envOpts.Segment.DBName = "dbnamefromenv"

	os.Setenv("GLEAFD_NAME", envOpts.Name)
	os.Setenv("GLEAFD_ADDR", envOpts.Addr)
	os.Setenv("GLEAFD_SEGMENT_DB_NAME", envOpts.Segment.DBName)

	parsedOpts, err := Load(args)
	if err != nil {
		t.Fatal(err)
	}
	// 命令行参数不存在，则使用环境变量
	if parsedOpts.Name != envOpts.Name {
		t.Fatalf("name = %v, want = %v", parsedOpts.Name, envOpts.Name)
	}
	// 命令行参数和环境变量同时存在，则使用命令行参数。
	if parsedOpts.Segment.DBName == envOpts.Segment.DBName {
		t.Fatalf("segment-db-name = %v, want = %v", parsedOpts.Segment.DBName, envOpts.Segment.DBName)
	}
	if parsedOpts.Addr != cliOpts.Addr {
		t.Fatalf("addr = %v, want = %v", parsedOpts.Addr, cliOpts.Addr)
	}
}

func TestParseFromFile(t *testing.T) {
	// TODO:
	cliRedisAddr := "abc.com:8490"
	args := []string{
		"--config=gleafd_test.yaml",
		"--snowflake-redis-addr=" + cliRedisAddr,
	}
	envName := "myGleafd"
	os.Setenv("GLEAFD_NAME", envName)

	opts, err := Load(args)
	if err != nil {
		t.Fatal(err)
	}
	// 所有配置从配置文件中读取
	if opts.Name != "gleafd" {
		t.Fatalf("name = %v, want = gleafd", opts.Name)
	}
}
