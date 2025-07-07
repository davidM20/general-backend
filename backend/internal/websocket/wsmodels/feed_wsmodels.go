package wsmodels

/*
 * ===================================================
 * MODELOS DE DATOS PARA EL FEED WEBSOCKET
 * ===================================================
 *
 * ESTRUCTURA DE DATOS:
 * --------------------
 * Estos modelos definen la estructura de los items que se envían a través de WebSocket
 * para el feed de la aplicación. Están diseñados para ser compatibles con lo que
 * espera el frontend.
 *
 * TIPOS DE ITEMS:
 * ---------------
 * - FeedItem:      La estructura principal para cualquier item del feed.
 * - StudentFeedData: Datos específicos para un item de tipo "student".
 * - CompanyFeedData: Datos específicos para un item de tipo "company".
 * - EventFeedData:   Datos específicos para un item de tipo "event".
 *
 * USO:
 * ----
 * Estos modelos son utilizados por FeedService para construir la respuesta
 * y por FeedHandler para enviarla al cliente.
 */

// FeedItem representa un elemento genérico en el feed.
type FeedItem struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`      // "student", "company", "event"
	Timestamp string      `json:"timestamp"` // Formato legible (ej: "2 hours ago")
	Data      interface{} `json:"data"`      // Contendrá StudentFeedData, CompanyFeedData, o EventFeedData
}

// StudentFeedData contiene los datos específicos para un item del feed de tipo "student".
type StudentFeedData struct {
	Name        string   `json:"name"`
	Avatar      string   `json:"avatar"` // URL
	Career      string   `json:"career"`
	University  string   `json:"university"`
	Skills      []string `json:"skills"`
	Description string   `json:"description"`
	UserID      int64    `json:"userId"`
	UserName    string   `json:"userName"`
}

// CompanyFeedData contiene los datos específicos para un item del feed de tipo "company".
type CompanyFeedData struct {
	Name        string `json:"name"`
	Logo        string `json:"logo"` // URL
	Industry    string `json:"industry"`
	Location    string `json:"location"`
	Description string `json:"description"`
	UserID      int64  `json:"userId"`
	UserName    string `json:"userName"`
}

// EventFeedData contiene los datos específicos para un item del feed de tipo "event".
type EventFeedData struct {
	Title       string `json:"title"`
	Company     string `json:"company"`     // Nombre de la compañía organizadora
	CompanyLogo string `json:"companyLogo"` // URL del logo de la compañía
	Date        string `json:"date"`        // Formato legible (ej: "Nov 15, 2023")
	Location    string `json:"location"`
	Image       string `json:"image"` // URL de la imagen del evento
	Description string `json:"description"`
	PostType    string `json:"postType"` // Diferenciar entre 'EVENTO', 'DESAFIO', 'ARTICULO', etc.
	EventID     int64  `json:"eventId"`
}

// PaginationInfo contiene detalles sobre la paginación de una lista.
type PaginationInfo struct {
	TotalItems  int  `json:"totalItems"`
	CurrentPage int  `json:"currentPage"`
	HasMore     bool `json:"hasMore"`
}

// FeedListResponsePayload es el payload para la respuesta de la lista de feed.
type FeedListResponsePayload struct {
	Items      []FeedItem      `json:"items"`
	Pagination *PaginationInfo `json:"pagination"`
}

// FeedItemViewRef es una referencia a un item del feed que ha sido visto.
type FeedItemViewRef struct {
	ItemType string `json:"itemType"` // 'user' o 'event'
	ItemID   int64  `json:"itemId"`
}
