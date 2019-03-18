package config

import (
	"os"
	"testing"

	"github.com/derry6/gleafd/server"
)

func b2s(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func getTestOptions() *server.Options {
	wantOpts := server.NewOptions()
	wantOpts.Name = "gleafd"
	wantOpts.Addr = "localhost:7890"
	wantOpts.Segment.Enable = false
	wantOpts.Segment.DBHost = "example.com:3306"
	wantOpts.Segment.DBName = "examaple"
	wantOpts.Segment.DBUser = "gleafd"
	wantOpts.Segment.DBPass = "12454656"
	wantOpts.Snowflake.Enable = false
	wantOpts.Snowflake.ZkAddress = "www.gleafd.com:2191"
	return wantOpts
}

func TestParseArgs(t *testing.T) {
	wantOpts := getTestOptions()
	args := []string{
		"name=" + wantOpts.Name,
		"addr=" + wantOpts.Addr,
		"--segment-enable=" + b2s(wantOpts.Segment.Enable),
		"--segment-db-host=" + wantOpts.Segment.DBHost,
		"--segment-db-user=" + wantOpts.Segment.DBUser,
		"--segment-db-pass=" + wantOpts.Segment.DBPass,
		"--segment-db-name=" + wantOpts.Segment.DBName,
		"--snowflake-enable=" + b2s(wantOpts.Snowflake.Enable),
		"--snowflake-zk-addr=" + wantOpts.Snowflake.ZkAddress,
	}
	cfg := New()
	if err := cfg.Parse(args); err != nil {
		t.Fatal(err)
	}
	opts := cfg.Options()
	if *opts != *wantOpts {
		t.Fatalf("opts = %v, want = %v", opts, wantOpts)
	}
}

func TestParseArgsWithEnv(t *testing.T) {
	cliOpts := getTestOptions()
	args := []string{
		// "name=" + cliOpts.Name,
		"addr=" + cliOpts.Addr,
		"--segment-enable=" + b2s(cliOpts.Segment.Enable),
		"--segment-db-host=" + cliOpts.Segment.DBHost,
		"--segment-db-user=" + cliOpts.Segment.DBUser,
		"--segment-db-pass=" + cliOpts.Segment.DBPass,
		"--segment-db-name=" + cliOpts.Segment.DBName,
		// "--snowflake-enable=" + b2s(cliOpts.Snowflake.Enable),
		"--snowflake-zk-addr=" + cliOpts.Snowflake.ZkAddress,
	}
	envOpts := server.NewOptions()
	envOpts.Name = "myGleafd"
	envOpts.Addr = "fromenv:1239"
	envOpts.Segment.DBName = "dbnamefromenv"

	os.Setenv("GLEAFD_SERVER_NAME", envOpts.Name)
	os.Setenv("GLEAFD_SERVER_ADDR", envOpts.Addr)
	os.Setenv("GLEAFD_SEGMENT_DB_NAME", envOpts.Segment.DBName)

	cfg := New()
	if err := cfg.Parse(args); err != nil {
		t.Fatal(err)
	}
	parsedOpts := cfg.Options()
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
	cliZkAddr := "abc.com:8490"
	args := []string{
		"--config=gleafd_test.yaml",
		"--snowflake-zk-addr=" + cliZkAddr,
	}
	envName := "myGleafd"
	os.Setenv("GLEAFD_SERVER_NAME", envName)

	cfg := New()
	if err := cfg.Parse(args); err != nil {
		t.Fatal(err)
	}
	// 所有配置从配置文件中读取
	opts := cfg.Options()
	if opts.Addr != "123.345.980.123:9091" {
		t.Fatalf("addr = %v, want = 123.345.980.123:9091", opts.Addr)
	}
	if opts.Name != "gleafd" {
		t.Fatalf("name = %v, want = gleafd", opts.Name)
	}
	if opts.Snowflake.ZkAddress != "www.gleafd.com:2181" {
		t.Fatalf("zkAddr = %v, want = %v", opts.Snowflake.ZkAddress, "www.gleafd.com:2181")
	}
}
