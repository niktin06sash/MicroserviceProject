package configs

type Config struct {
	Server  ServerConfig         `mapstructure:"server"`
	Session SessionServiceConfig `mapstructure:"session_service"`
	Routes  map[string]string    `mapstructure:"routes"`
	SSL     SSLConfig            `mapstructure:"ssl"`
}
type ServerConfig struct {
	Port string `mapstructure:"port"`
}
type SessionServiceConfig struct {
	GrpcAddress string `mapstructure:"grpc_address"`
}
type SSLConfig struct {
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}
