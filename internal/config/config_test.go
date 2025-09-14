package config

import (
	"flag"
	"os"
	"testing"

	dbCfg "github.com/alex-storchak/shortener/internal/db/config"
	handlerCfg "github.com/alex-storchak/shortener/internal/handler/config"
	loggerCfg "github.com/alex-storchak/shortener/internal/logger/config"
	repoCfg "github.com/alex-storchak/shortener/internal/repository/config"
	"github.com/stretchr/testify/suite"
)

const (
	envNameServerAddress = "SERVER_ADDRESS"
	envNameBaseURL       = "BASE_URL"
	envNameLogLevel      = "LOG_LEVEL"
	envFileStoragePath   = "FILE_STORAGE_PATH"
	envDatabaseDSN       = "DATABASE_DSN"
)

type ConfigTestSuite struct {
	suite.Suite
	origEnvs        map[string]string
	envVarsNames    []string
	origArgs        []string
	origCommandLine *flag.FlagSet
}

func (s *ConfigTestSuite) SetupSuite() {
	s.envVarsNames = []string{envNameServerAddress, envNameBaseURL, envNameLogLevel, envFileStoragePath, envDatabaseDSN}
}

func (s *ConfigTestSuite) setEnvs(envs map[string]string) {
	for name, val := range envs {
		if err := os.Setenv(name, val); err != nil {
			s.T().Fatal(err)
		}
	}
}

func (s *ConfigTestSuite) unsetEnvs(names []string) {
	for _, name := range names {
		if err := os.Unsetenv(name); err != nil {
			s.T().Fatal(err)
		}
	}
}

func (s *ConfigTestSuite) SetupSubTest() {
	for _, key := range s.envVarsNames {
		if val, exists := os.LookupEnv(key); exists {
			s.origEnvs[key] = val
		}
		if err := os.Unsetenv(key); err != nil {
			s.T().Fatal(err)
		}
	}
	s.origArgs = os.Args
	s.origCommandLine = flag.CommandLine
}

func (s *ConfigTestSuite) TearDownSubTest() {
	s.unsetEnvs(s.envVarsNames)
	for key, val := range s.origEnvs {
		if err := os.Setenv(key, val); err != nil {
			s.T().Fatal(err)
		}
	}
	s.origEnvs = make(map[string]string)
	os.Args = s.origArgs
	flag.CommandLine = s.origCommandLine
}

func (s *ConfigTestSuite) TestParseConfig() {
	dfltHandlerCfg := handlerCfg.Config{
		ServerAddr: handlerCfg.DefaultServerAddr,
		BaseURL:    handlerCfg.DefaultBaseURL,
	}
	dfltLoggerCfg := loggerCfg.Config{
		LogLevel: loggerCfg.DefaultLogLevel,
	}
	dfltRepoCfg := repoCfg.Config{
		FileStorage: repoCfg.FileStorage{
			Path: repoCfg.DefaultFileStoragePath,
		},
	}
	dfltDBCfg := dbCfg.Config{
		DSN: dbCfg.DefaultDatabaseDSN,
	}

	tests := []struct {
		name  string
		flags []string
		envs  map[string]string
		want  *Config
	}{
		{
			name:  "parse config without flags and envs returns default values",
			flags: []string{},
			envs:  map[string]string{},
			want: &Config{
				Handler:    dfltHandlerCfg,
				Logger:     dfltLoggerCfg,
				Repository: dfltRepoCfg,
				DB:         dfltDBCfg,
			},
		},
		{
			name: "parse config with custom flags without envs returns flags values",
			flags: []string{
				"-a=example.com:1111",
				"-b=http://example.com:1111",
				"-f=./data/some_file.json",
				"-d=postgres:flagsDSN",
			},
			envs: map[string]string{},
			want: &Config{
				Handler: handlerCfg.Config{
					ServerAddr: "example.com:1111",
					BaseURL:    "http://example.com:1111",
				},
				Logger: dfltLoggerCfg,
				Repository: repoCfg.Config{
					FileStorage: repoCfg.FileStorage{
						Path: "./data/some_file.json",
					},
				},
				DB: dbCfg.Config{DSN: "postgres:flagsDSN"},
			},
		},
		{
			name: "parse config with custom -a (server address) flag without envs returns -a flag value",
			flags: []string{
				"-a=example.com:1111",
			},
			envs: map[string]string{},
			want: &Config{
				Handler: handlerCfg.Config{
					ServerAddr: "example.com:1111",
					BaseURL:    handlerCfg.DefaultBaseURL,
				},
				Logger:     dfltLoggerCfg,
				Repository: dfltRepoCfg,
				DB:         dfltDBCfg,
			},
		},
		{
			name: "parse config with custom -b (short url base address) flag without envs returns -b flag value",
			flags: []string{
				"-b=http://example.com:1111",
			},
			envs: map[string]string{},
			want: &Config{
				Handler: handlerCfg.Config{
					ServerAddr: handlerCfg.DefaultServerAddr,
					BaseURL:    "http://example.com:1111",
				},
				Logger:     dfltLoggerCfg,
				Repository: dfltRepoCfg,
				DB:         dfltDBCfg,
			},
		},
		{
			name:  "parse config with custom envs without flags returns envs values",
			flags: []string{},
			envs: map[string]string{
				envNameServerAddress: "env-example.com:1111",
				envNameBaseURL:       "http://env-example.com:1111",
				envFileStoragePath:   "./data/some_file.json",
			},
			want: &Config{
				Handler: handlerCfg.Config{
					ServerAddr: "env-example.com:1111",
					BaseURL:    "http://env-example.com:1111",
				},
				Logger: dfltLoggerCfg,
				Repository: repoCfg.Config{
					FileStorage: repoCfg.FileStorage{
						Path: "./data/some_file.json",
					},
				},
				DB: dfltDBCfg,
			},
		},
		{
			name: "parse config with custom envs with flags returns envs values",
			flags: []string{
				"-a=flags-example.com:1111",
				"-b=http://flags-example.com:1111",
				"-f=./data/some_file.json",
			},
			envs: map[string]string{
				envNameServerAddress: "env-example.com:1111",
				envNameBaseURL:       "http://env-example.com:1111",
				envFileStoragePath:   "./data/some_another_file.json",
			},
			want: &Config{
				Handler: handlerCfg.Config{
					ServerAddr: "env-example.com:1111",
					BaseURL:    "http://env-example.com:1111",
				},
				Logger: dfltLoggerCfg,
				Repository: repoCfg.Config{
					FileStorage: repoCfg.FileStorage{
						Path: "./data/some_another_file.json",
					},
				},
				DB: dfltDBCfg,
			},
		},
		{
			name: "parse config with custom env SERVER_ADDRESS and -b (base url) flag returns env and flag values",
			flags: []string{
				"-b=http://flags-example.com:1111",
			},
			envs: map[string]string{
				envNameServerAddress: "env-example.com:1111",
			},
			want: &Config{
				Handler: handlerCfg.Config{
					ServerAddr: "env-example.com:1111",
					BaseURL:    "http://flags-example.com:1111",
				},
				Logger:     dfltLoggerCfg,
				Repository: dfltRepoCfg,
				DB:         dfltDBCfg,
			},
		},
		{
			name: "parse config with custom env SERVER_ADDRESS and -a (server address) flag returns env and default values",
			flags: []string{
				"-a=flags-example.com:1111",
			},
			envs: map[string]string{
				envNameServerAddress: "env-example.com:1111",
			},
			want: &Config{
				Handler: handlerCfg.Config{
					ServerAddr: "env-example.com:1111",
					BaseURL:    handlerCfg.DefaultBaseURL,
				},
				Logger:     dfltLoggerCfg,
				Repository: dfltRepoCfg,
				DB:         dfltDBCfg,
			},
		},
		{
			name: "parse config with custom env LOG_LEVEL and -l (log level) flag returns env",
			flags: []string{
				"-l=error",
			},
			envs: map[string]string{
				envNameLogLevel: "debug",
			},
			want: &Config{
				Handler: dfltHandlerCfg,
				Logger: loggerCfg.Config{
					LogLevel: "debug",
				},
				Repository: dfltRepoCfg,
				DB:         dfltDBCfg,
			},
		},
		{
			name:  "parse config with custom env LOG_LEVEL without flags returns env",
			flags: []string{},
			envs: map[string]string{
				envNameLogLevel: "debug",
			},
			want: &Config{
				Handler: dfltHandlerCfg,
				Logger: loggerCfg.Config{
					LogLevel: "debug",
				},
				Repository: dfltRepoCfg,
				DB:         dfltDBCfg,
			},
		},
		{
			name: "parse config without env with -l (log level) flag returns flag",
			flags: []string{
				"-l=error",
			},
			envs: map[string]string{},
			want: &Config{
				Handler: dfltHandlerCfg,
				Logger: loggerCfg.Config{
					LogLevel: "error",
				},
				Repository: dfltRepoCfg,
				DB:         dfltDBCfg,
			},
		},
		{
			name: "parse config with env FILE_STORAGE_PATH and -f (file storage path) flag returns env",
			flags: []string{
				"-f=./data/flags_file.json",
			},
			envs: map[string]string{
				envFileStoragePath: "./data/env_file.json",
			},
			want: &Config{
				Handler: dfltHandlerCfg,
				Logger:  dfltLoggerCfg,
				Repository: repoCfg.Config{
					FileStorage: repoCfg.FileStorage{
						Path: "./data/env_file.json",
					},
				},
				DB: dfltDBCfg,
			},
		},
		{
			name:  "parse config with env FILE_STORAGE_PATH without flags returns env",
			flags: []string{},
			envs: map[string]string{
				envFileStoragePath: "./data/env_file.json",
			},
			want: &Config{
				Handler: dfltHandlerCfg,
				Logger:  dfltLoggerCfg,
				Repository: repoCfg.Config{
					FileStorage: repoCfg.FileStorage{
						Path: "./data/env_file.json",
					},
				},
				DB: dfltDBCfg,
			},
		},
		{
			name: "parse config without env with -f (file storage path) flag returns flag",
			flags: []string{
				"-f=./data/flags_file.json",
			},
			envs: map[string]string{},
			want: &Config{
				Handler: dfltHandlerCfg,
				Logger:  dfltLoggerCfg,
				Repository: repoCfg.Config{
					FileStorage: repoCfg.FileStorage{
						Path: "./data/flags_file.json",
					},
				},
				DB: dfltDBCfg,
			},
		},
		{
			name: "parse config with env DATABASE_DSN and -d (database dsn) flag returns env",
			flags: []string{
				"-d=postgres:flagsDSN",
			},
			envs: map[string]string{
				envDatabaseDSN: "postgres:envDSN",
			},
			want: &Config{
				Handler:    dfltHandlerCfg,
				Logger:     dfltLoggerCfg,
				Repository: dfltRepoCfg,
				DB:         dbCfg.Config{DSN: "postgres:envDSN"},
			},
		},
		{
			name:  "parse config with env DATABASE_DSN without flags returns env",
			flags: []string{},
			envs: map[string]string{
				envDatabaseDSN: "postgres:envDSN",
			},
			want: &Config{
				Handler:    dfltHandlerCfg,
				Logger:     dfltLoggerCfg,
				Repository: dfltRepoCfg,
				DB:         dbCfg.Config{DSN: "postgres:envDSN"},
			},
		},
		{
			name: "parse config without env with -d (database dsn) flag returns flag",
			flags: []string{
				"-d=postgres:flagsDSN",
			},
			envs: map[string]string{},
			want: &Config{
				Handler:    dfltHandlerCfg,
				Logger:     dfltLoggerCfg,
				Repository: dfltRepoCfg,
				DB:         dbCfg.Config{DSN: "postgres:flagsDSN"},
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.setEnvs(tt.envs)
			flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
			testArgs := append([]string{"test"}, tt.flags...)
			os.Args = testArgs

			got, err := ParseConfig()

			s.Require().NoError(err)
			s.Assert().Equal(tt.want, got)
		})
	}
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
