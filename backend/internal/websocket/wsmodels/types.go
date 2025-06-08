package wsmodels

import "time"

// WsUserData se asocia con cada conexión WebSocket gestionada por customws.
// Contiene la información esencial del usuario para la sesión WebSocket.
type WsUserData struct {
	UserID   int64
	Username string
	RoleId   int
	// Podríamos añadir más datos aquí si son frecuentemente necesarios
	// y queremos evitar consultas repetidas a la BD en cada mensaje.
	// Por ejemplo: Roles, Email.
	// IsMyMessage   int    `json:"isMyMessage,omitempty"` // El frontend lo calcula, pero podría venir del backend
}

// ChatInfo representa la información resumida de un chat para la lista de chats del usuario.
type ChatInfo struct {
	ChatID                string `json:"chatId"`                          // Identificador único del chat (puede ser el ID del contacto si es un chat 1-a-1)
	OtherUserID           int64  `json:"otherUserId"`                     // ID del otro usuario en el chat
	OtherUserName         string `json:"otherUserName"`                   // Nombre de usuario del otro participante
	OtherFirstName        string `json:"otherFirstName,omitempty"`        // Nombre del otro participante
	OtherLastName         string `json:"otherLastName,omitempty"`         // Apellido del otro participante
	OtherPicture          string `json:"otherPicture,omitempty"`          // URL de la imagen de perfil del otro participante
	LastMessage           string `json:"lastMessage,omitempty"`           // Contenido del último mensaje en el chat
	LastMessageTs         int64  `json:"lastMessageTs,omitempty"`         // Timestamp Unix (en milisegundos) del último mensaje
	LastMessageFromUserId int64  `json:"lastMessageFromUserId,omitempty"` // ID del usuario que envió el último mensaje
	UnreadCount           int    `json:"unreadCount,omitempty"`           // Número de mensajes no leídos por el usuario actual en este chat
	IsOtherOnline         bool   `json:"isOnline"`                        // Estado de conexión del otro usuario
}

// NotificationInfo representa una notificación para el usuario.
// Se adapta a varios tipos de eventos dentro de la aplicación.
type NotificationInfo struct {
	ID             string      `json:"id"`                       // ID único del evento/notificación (proveniente de la tabla Event)
	Type           string      `json:"type"`                     // EventType de la tabla Event (FRIEND_REQUEST, SYSTEM, etc.)
	Title          string      `json:"title"`                    // EventTitle de la tabla Event
	Message        string      `json:"message"`                  // Description de la tabla Event
	Timestamp      time.Time   `json:"timestamp"`                // CreateAt de la tabla Event
	IsRead         bool        `json:"isRead"`                   // IsRead de la tabla Event
	Payload        interface{} `json:"payload,omitempty"`        // Metadata (parseada) de la tabla Event y otros IDs como otherUserId, projectId, groupId
	Profile        ProfileData `json:"profile,omitempty"`        // Información del OtherUserId (si existe)
	Status         string      `json:"status,omitempty"`         // Status de la tabla Event (PENDING, ACCEPTED, etc.)
	ActionRequired bool        `json:"actionRequired,omitempty"` // ActionRequired de la tabla Event
	ActionTakenAt  *time.Time  `json:"actionTakenAt,omitempty"`  // ActionTakenAt de la tabla Event (puede ser nil)
	OtherUserId    int64       `json:"otherUserId,omitempty"`    // OtherUserId de la tabla Event (directamente)
	ProyectId      int64       `json:"proyectId,omitempty"`      // ProyectId de la tabla Event (directamente)
	GroupId        int64       `json:"groupId,omitempty"`        // GroupId de la tabla Event (directamente)
}

// ProfileData representa la información completa del perfil de un usuario.
// Agrega datos de múltiples tablas de la base dedatos.
type ProfileData struct {
	ID                 int64           `json:"id"`
	FirstName          string          `json:"firstName"`
	LastName           string          `json:"lastName"`
	UserName           string          `json:"userName"`
	Email              string          `json:"email"`
	Phone              string          `json:"phone,omitempty"`
	Sex                string          `json:"sex,omitempty"`
	DocId              string          `json:"docId,omitempty"`
	NationalityId      int             `json:"nationalityId,omitempty"`
	NationalityName    string          `json:"nationalityName,omitempty"`
	Birthdate          string          `json:"birthdate,omitempty"` // Formato YYYY-MM-DD
	Picture            string          `json:"picture,omitempty"`
	DegreeName         string          `json:"degreeName,omitempty"`
	UniversityName     string          `json:"universityName,omitempty"`
	RoleID             int             `json:"roleId"`
	RoleName           string          `json:"roleName"`
	StatusAuthorizedId int             `json:"statusAuthorizedId"`
	Summary            string          `json:"summary,omitempty"`
	Address            string          `json:"address,omitempty"`
	Github             string          `json:"github,omitempty"`
	Linkedin           string          `json:"linkedin,omitempty"`
	CreatedAt          time.Time       `json:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt"`
	Curriculum         CurriculumVitae `json:"curriculum"`
	IsOnline           bool            `json:"isOnline,omitempty"`
}

// CurriculumVitae agrupa las secciones del currículum de un usuario.
type CurriculumVitae struct {
	Education      []EducationItem      `json:"education"`
	Experience     []WorkExperienceItem `json:"experience"`
	Certifications []CertificationItem  `json:"certifications"`
	Skills         []SkillItem          `json:"skills"`
	Languages      []LanguageItem       `json:"languages"`
	Projects       []ProjectItem        `json:"projects"`
}

// EducationItem representa una entrada en la sección de educación del currículum.
type EducationItem struct {
	ID                  int64  `json:"id"`
	Institution         string `json:"institution"`
	Degree              string `json:"degree"`
	Campus              string `json:"campus,omitempty"`
	GraduationDate      string `json:"graduationDate,omitempty"` // Formato YYYY-MM-DD
	CountryID           int64  `json:"countryId,omitempty"`      // Referencia a la tabla Nationality
	CountryName         string `json:"countryName,omitempty"`    // Nombre del país
	IsCurrentlyStudying bool   `json:"isCurrentlyStudying,omitempty"`
}

// WorkExperienceItem representa una entrada en la sección de experiencia laboral.
type WorkExperienceItem struct {
	ID           int64  `json:"id"`
	PersonId     int64  `json:"personId,omitempty"`
	Company      string `json:"company"`
	Position     string `json:"position"`
	StartDate    string `json:"startDate,omitempty"` // Formato YYYY-MM-DD
	EndDate      string `json:"endDate,omitempty"`   // Formato YYYY-MM-DD, puede ser nulo (actual)
	Description  string `json:"description,omitempty"`
	CountryID    int64  `json:"countryId,omitempty"`   // Referencia a la tabla Nationality
	CountryName  string `json:"countryName,omitempty"` // Nombre del país
	IsCurrentJob bool   `json:"isCurrentJob,omitempty"`
}

// CertificationItem representa una certificación obtenida por el usuario.
type CertificationItem struct {
	ID              int64  `json:"id"`
	Certification   string `json:"certification"`
	Institution     string `json:"institution"`
	DateObtained    string `json:"dateObtained,omitempty"`    // Formato YYYY-MM-DD
	ProjectStatus   string `json:"projectStatus,omitempty"`   // Ej: "En curso", "Completado"
	StartDate       string `json:"startDate,omitempty"`       // Formato YYYY-MM-DD
	ExpectedEndDate string `json:"expectedEndDate,omitempty"` // Formato YYYY-MM-DD
	IsOngoing       bool   `json:"isOngoing,omitempty"`
}

// SkillItem representa una habilidad del usuario.
type SkillItem struct {
	ID    int64  `json:"id"`
	Skill string `json:"skill"`
	Level string `json:"level"` // ej: "Principiante", "Intermedio", "Avanzado", "Experto"
}

// LanguageItem representa un idioma que el usuario conoce.
type LanguageItem struct {
	ID       int64  `json:"id"`
	Language string `json:"language"`
	Level    string `json:"level"` // ej: "A1", "B2", "Nativo"
}

// ProjectItem representa un proyecto en el que el usuario ha participado.
type ProjectItem struct {
	ID              int64  `json:"id"`
	Title           string `json:"title"`
	Role            string `json:"role"`
	Description     string `json:"description,omitempty"`
	Company         string `json:"company,omitempty"`         // Empresa asociada (si aplica)
	Document        string `json:"document,omitempty"`        // URL o referencia a documentación/evidencia
	ProjectStatus   string `json:"projectStatus,omitempty"`   // Ej: "En curso", "Completado"
	StartDate       string `json:"startDate,omitempty"`       // Formato YYYY-MM-DD
	ExpectedEndDate string `json:"expectedEndDate,omitempty"` // Formato YYYY-MM-DD
	IsOngoing       bool   `json:"isOngoing,omitempty"`
}

// UserContactInfo se utiliza para mostrar información de usuarios en listas de contactos o resultados de búsqueda.
type UserContactInfo struct {
	ID        int64  `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	UserName  string `json:"userName"`
	Picture   string `json:"picture,omitempty"`
	Summary   string `json:"summary,omitempty"` // Un breve resumen o "estado" del usuario
	IsOnline  bool   `json:"isOnline"`          // Si el usuario está actualmente conectado al WebSocket
}

// EnterpriseInfo se utiliza para mostrar información de empresas en resultados de búsqueda.
type EnterpriseInfo struct {
	ID           int64  `json:"id"`
	RIF          string `json:"rif"`                    // Identificador fiscal de la empresa
	CompanyName  string `json:"companyName"`            // Nombre de la empresa
	CategoryID   int64  `json:"categoryId,omitempty"`   // ID de la categoría de la empresa
	CategoryName string `json:"categoryName,omitempty"` // Nombre de la categoría
	Description  string `json:"description,omitempty"`
	Location     string `json:"location,omitempty"`
	Phone        string `json:"phone,omitempty"`
	// Se podría añadir un logo o imagen de la empresa
}

// MessageDB representa la estructura de un mensaje como se almacena en la base de datos,
// alineada con la tabla 'Message' del schema.sql robusto.
type MessageDB struct {
	Id               string  `json:"id"`                         // ID único del mensaje (UUID).
	ChatId           *string `json:"chatId,omitempty"`           // ID del chat privado, nulo si es un mensaje de grupo.
	ChatIdGroup      *string `json:"chatIdGroup,omitempty"`      // ID del grupo, nulo si es un mensaje privado.
	SenderId         int64   `json:"senderId"`                   // ID del usuario que envió el mensaje.
	TypeMessageId    int64   `json:"typeMessageId"`              // ID del tipo de mensaje.
	Content          *string `json:"content,omitempty"`          // Contenido de texto del mensaje, nulo si es solo multimedia.
	MediaId          *string `json:"mediaId,omitempty"`          // ID del archivo multimedia adjunto, nulo si es solo texto.
	ReplyToMessageId *string `json:"replyToMessageId,omitempty"` // ID del mensaje al que se responde.
	SentAt           string  `json:"sentAt"`                     // Timestamp ISO8601 UTC del envío.
	EditedAt         *string `json:"editedAt,omitempty"`         // Timestamp ISO8601 UTC de la última edición.
	Status           string  `json:"status"`                     // Estado: 'sending', 'sent', 'delivered', 'read', 'failed'.
}

// WsMessage es una estructura genérica para los mensajes WebSocket salientes.
// Type indica el tipo de mensaje (ej: "chat_message", "notification", "user_status")
// Payload contiene los datos específicos del mensaje.
type WsMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// DashboardDataPayload representa los datos para el dashboard de administración.
type DashboardDataPayload struct {
	ActiveUsers          int64           `json:"activeUsers"`
	TotalRegisteredUsers int64           `json:"totalRegisteredUsers"`
	AdministrativeUsers  int64           `json:"administrativeUsers"`
	BusinessAccounts     int64           `json:"businessAccounts"`
	AlumniStudents       int64           `json:"alumniStudents"`
	EgresadoUsers        int64           `json:"egresadoUsers"`
	AverageUsageTime     string          `json:"averageUsageTime"` // Formato "Xh Ym" o similar
	UsersByCampus        []UserByCampus  `json:"usersByCampus"`
	MonthlyActivity      MonthlyActivity `json:"monthlyActivity"`
}

// UserByCampus representa el número de usuarios por campus.
type UserByCampus struct {
	Users int64  `json:"users"`
	Name  string `json:"name"` // Nombre del campus
}

// MonthlyActivity representa la actividad mensual.
type MonthlyActivity struct {
	Labels []string `json:"labels"` // Ej: ["Ene", "Feb", "Mar"]
	Data   []int64  `json:"data"`
}

// Tipos de mensaje para el CV
const (
	MessageTypeSetSkill          = "set_skill"
	MessageTypeSetLanguage       = "set_language"
	MessageTypeSetWorkExperience = "set_work_experience"
	MessageTypeSetCertification  = "set_certification"
	MessageTypeSetProject        = "set_project"
	MessageTypeGetCV             = "get_cv"
)
