package models

// EnterpriseRegistration representa la información mínima para registrar una empresa
type EnterpriseRegistration struct {
	CompanyName string `json:"companyName"`
	RIF         string `json:"rif"`
	Sector      string `json:"sector"`
	FirstName   string `json:"contactName"` // Nombre del contacto
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Password    string `json:"password,omitempty"` // No se muestra en las respuestas
	Location    string `json:"location,omitempty"`
}

// EnterpriseResponse representa la respuesta después de registrar una empresa
type EnterpriseResponse struct {
	ID          int64  `json:"id"`
	CompanyName string `json:"companyName"`
	RIF         string `json:"rif"`
	Email       string `json:"email"`
	Message     string `json:"message,omitempty"`
}
