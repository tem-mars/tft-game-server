package app

type Config struct {
    Server struct {
        Host string
        Port string
    }
    JWT struct {
        Secret string
    }
}

func LoadConfig() (*Config, error) {
    cfg := &Config{}
    
    cfg.Server.Host = "localhost"
    cfg.Server.Port = "8080"
    cfg.JWT.Secret = "your-secret-key"  

    return cfg, nil
}