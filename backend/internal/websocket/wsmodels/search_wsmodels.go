package wsmodels

// SearchResultItem representa un único resultado de búsqueda, que puede ser un usuario o una empresa.
type SearchResultItem struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`      // "user", "company", "graduate"
	Timestamp string      `json:"timestamp"` // Puede ser la fecha de creación/actualización
	Data      interface{} `json:"data"`      // Contendrá UserSearchResultData o CompanySearchResultData
}

// UserSearchResultData contiene los datos específicos para un resultado de búsqueda de tipo "user" o "graduate".
type UserSearchResultData struct {
	Name       string `json:"name"`
	Avatar     string `json:"avatar"` // URL
	Career     string `json:"career,omitempty"`
	University string `json:"university,omitempty"`
	Headline   string `json:"headline"`
}

// CompanySearchResultData contiene los datos específicos para un resultado de búsqueda de tipo "company".
type CompanySearchResultData struct {
	Name     string `json:"name"`
	Logo     string `json:"logo"` // URL
	Industry string `json:"industry"`
	Location string `json:"location"`
	Headline string `json:"headline"`
}
