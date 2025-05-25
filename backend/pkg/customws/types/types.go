package types

// Apiresponse es una estructura de respuesta genérica.
type Apiresponse struct {
	ApiOrigin string `json:"apiOrigin"`
	Status    string `json:"status"`
}
