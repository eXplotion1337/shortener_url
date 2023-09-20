package config

import (
	"flag"
	"net"
	"os"
)

type Config struct {
	StoragePath string
	ServerAddr  string
	BaseURL     string
	DataBaseDSN string
	TypeStorage string
}

type Builder struct {
	config *Config
}

func NewConfigBuilder() *Builder {
	return &Builder{
		config: &Config{},
	}
}

func (b *Builder) Storage(storagePath string) *Builder {
	b.config.StoragePath = storagePath
	return b
}

func (b *Builder) Address(serverAddr string) *Builder {
	b.config.ServerAddr = serverAddr
	return b
}

func (b *Builder) BaseURL(baseURL string) *Builder {
	b.config.BaseURL = baseURL
	return b
}

func (b *Builder) DataBase(dataBase string) *Builder {
	b.config.DataBaseDSN = dataBase
	return b
}

func (b *Builder) TypeStorage(TypeStorage string) *Builder {
	b.config.TypeStorage = TypeStorage
	return b
}

func (b *Builder) Build() *Config {
	return b.config
}

func getEnvOrFlag(envKey string, flagValue string, defaultValue string) string {
	envVal := os.Getenv(envKey)
	if envVal == "" && flagValue != "" {
		return flagValue
	} else if envVal == "" {
		return defaultValue
	}
	return envVal
}

func InitConfig() (*Config, error) {
	var (
		addrFlag     string
		baseURLFlag  string
		fileFlag     string
		dataBaseFlag string
		typeStor     string
	)

	flag.StringVar(&addrFlag, "a", "", "HTTP-сервера")
	flag.StringVar(&baseURLFlag, "b", "", "Базовый адрес результирующего сокращённого URL")
	flag.StringVar(&fileFlag, "f", "", "Путь до файла с сокращёнными URL")
	flag.StringVar(&dataBaseFlag, "d", "", "Подключение к базе данных")
	flag.Parse()

	serverAddress := getEnvOrFlag("SERVER_ADDRESS", addrFlag, "127.0.0.1:8080")
	baseURL := getEnvOrFlag("BASE_URL", baseURLFlag, "http://127.0.0.1:8080")
	fileStorage := getEnvOrFlag("FILE_STORAGE_PATH", fileFlag, "./")
	dataBaseDsn := getEnvOrFlag("DATABASE_DSN", dataBaseFlag, "")

	_, err := net.ResolveTCPAddr("tcp", serverAddress)
	if err != nil {
		serverAddress = "127.0.0.1:8080"
		baseURL = "http://127.0.0.1:8080"
	}

	if dataBaseDsn == "" {
		if fileStorage == "./" {
			typeStor = "In-memoryStorage"
		} else {
			typeStor = "FileStorage"
		}
	} else {
		typeStor = "DataBaseStorage"
	}

	builder := NewConfigBuilder().
		Address(serverAddress).
		BaseURL(baseURL).
		Storage(fileStorage).
		DataBase(dataBaseDsn).
		TypeStorage(typeStor)

	return builder.Build(), nil
}
