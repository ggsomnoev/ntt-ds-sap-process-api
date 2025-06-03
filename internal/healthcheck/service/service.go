package service

type Component interface {
	Name() string
	Check() error
}

type HealthCheckService struct {
	components []Component
}

func NewHealthCheckService(components ...Component) *HealthCheckService {
	return &HealthCheckService{components: components}
}

func (h *HealthCheckService) Status() (map[string]string, bool) {
	status := make(map[string]string)
	allHealthy := true

	for _, c := range h.components {
		if err := c.Check(); err != nil {
			status[c.Name()] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			status[c.Name()] = "ok"
		}
	}

	return status, allHealthy
}