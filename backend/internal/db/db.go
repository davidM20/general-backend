package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/davidM20/micro-service-backend-go.git/internal/models" // Ajusta la ruta si es necesario
	"github.com/davidM20/micro-service-backend-go.git/pkg/logger"
	"github.com/go-sql-driver/mysql"
)

var (
	db   *sql.DB
	once sync.Once
)

// Connect initializes the database connection.
// It expects the DSN (Data Source Name) for the MySQL database.
func Connect(dsn string) (*sql.DB, error) {
	var err error

	// Parse the DSN using the MySQL driver's parser
	cfg, err := mysql.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("could not parse DSN: %w", err)
	}

	dbName := cfg.DBName
	// Temporarily remove the database name to connect to the server
	cfg.DBName = ""
	serverDSN := cfg.FormatDSN()

	// Open connection to the database server
	serverDB, err := sql.Open("mysql", serverDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database server: %w", err)
	}
	defer serverDB.Close()

	// Create the database if it doesn't exist
	_, err = serverDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", dbName))
	if err != nil {
		return nil, fmt.Errorf("failed to create database '%s': %w", dbName, err)
	}

	// Now, connect to the specific database
	once.Do(func() {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			return // err will be handled outside once.Do
		}

		db.SetConnMaxLifetime(time.Minute * 3)
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(10)

		err = db.Ping()
		if err != nil {
			db.Close()         // Close the connection if ping fails
			db = nil           // Reset db to nil so subsequent calls can retry
			once = sync.Once{} // Reset once so connection can be retried
			return
		}
		logger.Success("DB", "Database connection successful!")
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	if db == nil { // Check if connection failed inside once.Do
		return nil, fmt.Errorf("database connection failed and was reset")
	}
	return db, nil
}

// GetDB returns the existing database connection pool.
// It's recommended to call Connect first.
func GetDB() *sql.DB {
	if db == nil {
		logger.Warn("DB", "Warning: GetDB called before Connect or connection failed.")
		// Depending on the strategy, you might want to panic or attempt reconnection here.
	}
	return db
}

// InitializeDatabase creates tables if they don't exist and populates default data.
func InitializeDatabase(conn *sql.DB) error {
	if conn == nil {
		return fmt.Errorf("database connection is nil")
	}

	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if anything fails

	if err := createTables(tx); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	if err := insertDefaultData(tx); err != nil {
		return fmt.Errorf("failed to insert default data: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Success("DB", "Database initialized successfully.")
	return nil
}

// createTables executes the CREATE TABLE IF NOT EXISTS statements.
func createTables(tx *sql.Tx) error {
	sqlSchema := `
    CREATE TABLE IF NOT EXISTS Token (
        Id INT PRIMARY KEY,
        TokenType VARCHAR(255) UNIQUE NOT NULL
    );

    CREATE TABLE IF NOT EXISTS Category (
        CategoryId BIGINT AUTO_INCREMENT PRIMARY KEY,
        Name VARCHAR(255),
        Description VARCHAR(255)
    );

    CREATE TABLE IF NOT EXISTS Interest (
        InterestId BIGINT AUTO_INCREMENT PRIMARY KEY,
        CategoryId BIGINT,
        Description VARCHAR(255),
        ExperienceLevel VARCHAR(255),
        FOREIGN KEY (CategoryId) REFERENCES Category(CategoryId)
    );

    CREATE TABLE IF NOT EXISTS TypeMessage (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        Name VARCHAR(255),
        Description VARCHAR(255)
    );

    CREATE TABLE IF NOT EXISTS Nationality (
        Id INT AUTO_INCREMENT PRIMARY KEY,
        CountryName VARCHAR(255) UNIQUE,
        IsoCode VARCHAR(255),
        DocIdFormat VARCHAR(255)
    );

    CREATE TABLE IF NOT EXISTS StatusAuthorized (
        Id INT PRIMARY KEY,
        Name VARCHAR(255) UNIQUE
    );

    CREATE TABLE IF NOT EXISTS University (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        Name VARCHAR(255) UNIQUE,
        Campus VARCHAR(255)
    );

    CREATE TABLE IF NOT EXISTS Degree (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        DegreeName VARCHAR(255),
        Descriptions VARCHAR(255),
        Code VARCHAR(255),
        UniversityId BIGINT,
        FOREIGN KEY (UniversityId) REFERENCES University(Id)
    );

    CREATE TABLE IF NOT EXISTS Role (
        Id INT PRIMARY KEY,
        Name VARCHAR(255) UNIQUE
    );


/*
Tabla User
Descripción: Esta tabla almacena la información tanto de usuarios individuales como de empresas.
La distinción entre tipo de usuario se maneja a través del campo RoleId.
Para usuarios individuales: Se utilizan los campos personales (FirstName, LastName, etc.)
Para empresas: Se utilizan los campos empresariales (RIF, CompanyName, Sector, etc.)

Campos principales:
- Información personal: FirstName, LastName, Email, Phone, etc.
- Información empresarial: RIF, CompanyName, Sector, Location, etc.
- Información de contacto: Email, ContactEmail, Phone, Address
- Redes sociales: Github, Linkedin, Twitter, Facebook
- Información académica: DegreeId, UniversityId
- Información de estado: RoleId, StatusAuthorizedId

Notas:
- El campo Email es único y obligatorio para todos los usuarios
- El campo RIF es único y obligatorio solo para empresas
- Los timestamps (CreatedAt, UpdatedAt) se actualizan automáticamente
*/

    CREATE TABLE IF NOT EXISTS User (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        FirstName VARCHAR(255),
        LastName VARCHAR(255),
        UserName VARCHAR(255) UNIQUE,
        Password VARCHAR(255),
        Email VARCHAR(255) UNIQUE NOT NULL,
ContactEmail VARCHAR(255),
Twitter VARCHAR(255),
Facebook VARCHAR(255),
        Phone VARCHAR(255),
        Sex VARCHAR(255),
        DocId VARCHAR(255) UNIQUE,
        NationalityId INT,
        Birthdate DATE,
        Picture VARCHAR(255),
DegreeId BIGINT, -- desusado
UniversityId BIGINT, -- desusado
RoleId INT,  -- el rol determina si es un estudiante o una empresa (1: estudiante, 2: egresado 3: empresa)
        StatusAuthorizedId INT,
Summary TEXT,
        Address VARCHAR(255),
        Github VARCHAR(255),
        Linkedin VARCHAR(255),
RIF VARCHAR(20) UNIQUE,
Sector VARCHAR(100),
CompanyName VARCHAR(255),
Location VARCHAR(255),
FoundationYear INT,
EmployeeCount INT,
dmeta_person_primary VARCHAR(24) NOT NULL DEFAULT '',
dmeta_person_secondary VARCHAR(24) NOT NULL DEFAULT '',
dmeta_company_primary VARCHAR(24) NOT NULL DEFAULT '',
dmeta_company_secondary VARCHAR(24) NOT NULL DEFAULT '',
CreatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
UpdatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
        FOREIGN KEY (NationalityId) REFERENCES Nationality(Id),
        FOREIGN KEY (DegreeId) REFERENCES Degree(Id),
        FOREIGN KEY (UniversityId) REFERENCES University(Id),
        FOREIGN KEY (RoleId) REFERENCES Role(Id),
        FOREIGN KEY (StatusAuthorizedId) REFERENCES StatusAuthorized(Id)
    );

    CREATE TABLE IF NOT EXISTS Online (
        UserOnlineId BIGINT PRIMARY KEY,
CreateAt DATE,
        Status TINYINT(1),
FOREIGN KEY (UserOnlineId) REFERENCES User(Id)
);

CREATE TABLE IF NOT EXISTS Contact (
ContactId BIGINT AUTO_INCREMENT PRIMARY KEY,
User1Id BIGINT,
User2Id BIGINT,
Status VARCHAR(255), --  'pending', 'accepted', 'rejected'
ChatId VARCHAR(255) UNIQUE,
FOREIGN KEY (User1Id) REFERENCES User(Id),
FOREIGN KEY (User2Id) REFERENCES User(Id)
);

CREATE TABLE IF NOT EXISTS GroupsUsers (
		Id BIGINT AUTO_INCREMENT PRIMARY KEY,
Name VARCHAR(255),
Description VARCHAR(255),
Picture VARCHAR(255),
AdminOfGroup BIGINT,
ChatId VARCHAR(255) UNIQUE,
FOREIGN KEY (AdminOfGroup) REFERENCES User(Id)
	);

    CREATE TABLE IF NOT EXISTS Multimedia (
    Id VARCHAR(255) PRIMARY KEY,
        Type VARCHAR(255),
        Ratio FLOAT,
        UserId BIGINT,
        FileName VARCHAR(255),
    CreateAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ContentId VARCHAR(255),
    ChatId VARCHAR(255),
    Size BIGINT,
    ProcessingStatus VARCHAR(50),
    Duration FLOAT,
    HLSManifestBaseURL VARCHAR(255),
    HLSManifest1080p VARCHAR(255),
    HLSManifest720p VARCHAR(255),
    HLSManifest480p VARCHAR(255)
    );

    CREATE TABLE IF NOT EXISTS Session (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        UserId BIGINT,
Tk VARCHAR(255),
        Ip VARCHAR(255),
        RoleId INT,
        TokenId INT,
FOREIGN KEY (UserId) REFERENCES User(Id),
FOREIGN KEY (RoleId) REFERENCES Role(Id)
);

/*
Tabla Message (versión robusta)
Descripción: Almacena todos los mensajes, tanto en chats privados como en grupos.

Mejoras sobre la versión original:
- Id: Se mantiene como VARCHAR(255) para soportar UUIDs generados por el cliente. Se recomienda usar CHAR(36) si son UUIDs estándar para ahorrar espacio y mejorar rendimiento.
- Semántica de nombres: Se han renombrado campos como UserId a SenderId y ResponseTo a ReplyToMessageId para mayor claridad.
- Contenido del mensaje: Text se cambia a Content y su tipo a TEXT para permitir mensajes más largos.
- Timestamps precisos: Date (que solo guardaba la fecha) se reemplaza por SentAt (DATETIME) para incluir la hora y se añade EditedAt para registrar ediciones.
- Estado del mensaje: StatusMessage (INT) se convierte en un ENUM para que los valores sean auto-descriptivos ('sending', 'sent', 'delivered', 'read', 'failed').
- Integridad de datos: Se añaden restricciones (CHECK constraints) para:
    1. Asegurar que un mensaje pertenezca a un chat (ChatId) O a un grupo (ChatIdGroup), pero no a ambos.
    2. Evitar mensajes vacíos (debe tener Content o MediaId).
- Índices optimizados: Se mueven los índices aquí y se ajustan para consultas comunes.
*/
    CREATE TABLE IF NOT EXISTS Message (
    Id VARCHAR(255) PRIMARY KEY,
    -- El ChatId o ChatIdGroup no puede ser nulo, pero solo uno de ellos debe tener valor.
    ChatId VARCHAR(255),
    ChatIdGroup VARCHAR(255),

    SenderId BIGINT NOT NULL,
    TypeMessageId BIGINT NOT NULL,
    
    Content TEXT,
        MediaId VARCHAR(255),
    
    -- Para mensajes que son una respuesta a otro.
    ReplyToMessageId VARCHAR(255),

    SentAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    EditedAt DATETIME, -- Se actualiza si el mensaje es editado.

    Status ENUM('sending', 'sent', 'delivered', 'read', 'failed') NOT NULL DEFAULT 'sending',

    FOREIGN KEY (SenderId) REFERENCES User(Id),
    FOREIGN KEY (TypeMessageId) REFERENCES TypeMessage(Id),
    FOREIGN KEY (MediaId) REFERENCES Multimedia(Id),
    FOREIGN KEY (ChatId) REFERENCES Contact(ChatId),
    FOREIGN KEY (ChatIdGroup) REFERENCES GroupsUsers(ChatId),
    FOREIGN KEY (ReplyToMessageId) REFERENCES Message(Id),
    
    -- Un mensaje debe tener contenido de texto o un adjunto.
    CONSTRAINT chk_message_content CHECK (Content IS NOT NULL OR MediaId IS NOT NULL),
    
    -- Un mensaje pertenece a un chat privado o a un grupo, no a ambos ni a ninguno.
    CONSTRAINT chk_message_chat_or_group CHECK (
        (ChatId IS NOT NULL AND ChatIdGroup IS NULL) OR 
        (ChatId IS NULL AND ChatIdGroup IS NOT NULL)
    )
);


CREATE TABLE IF NOT EXISTS GroupMembers (
        UserId BIGINT,
GroupId BIGINT,
FOREIGN KEY (UserId) REFERENCES User(Id),
FOREIGN KEY (GroupId) REFERENCES GroupsUsers(Id)
    );

    CREATE TABLE IF NOT EXISTS Education (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        PersonId BIGINT,
        Institution VARCHAR(255),
        Degree VARCHAR(255),
        Campus VARCHAR(255),
        GraduationDate DATE,
CountryId BIGINT,
IsCurrentlyStudying BOOLEAN DEFAULT FALSE,
dmeta_institution_primary VARCHAR(24) NOT NULL DEFAULT '',
dmeta_institution_secondary VARCHAR(24) NOT NULL DEFAULT '',
dmeta_degree_primary VARCHAR(24) NOT NULL DEFAULT '',
dmeta_degree_secondary VARCHAR(24) NOT NULL DEFAULT '',
FOREIGN KEY (PersonId) REFERENCES User(Id)
);


    CREATE TABLE IF NOT EXISTS WorkExperience (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        PersonId BIGINT,
        Company VARCHAR(255),
        Position VARCHAR(255),
        StartDate DATE,
        EndDate DATE,
Description TEXT,
CountryId BIGINT,
IsCurrentJob BOOLEAN DEFAULT FALSE,
dmeta_company_primary VARCHAR(24) NOT NULL DEFAULT '',
dmeta_company_secondary VARCHAR(24) NOT NULL DEFAULT '',
dmeta_position_primary VARCHAR(24) NOT NULL DEFAULT '',
dmeta_position_secondary VARCHAR(24) NOT NULL DEFAULT '',
FOREIGN KEY (PersonId) REFERENCES User(Id)
);


    CREATE TABLE IF NOT EXISTS Certifications (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        PersonId BIGINT,
        Certification VARCHAR(255),
        Institution VARCHAR(255),
        DateObtained DATE,
FOREIGN KEY (PersonId) REFERENCES User(Id)
    );

    CREATE TABLE IF NOT EXISTS Skills (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        PersonId BIGINT,
        Skill VARCHAR(255),
        Level VARCHAR(255),
dmeta_primary VARCHAR(12) NOT NULL DEFAULT '',
dmeta_secondary VARCHAR(12) NOT NULL DEFAULT '',
FOREIGN KEY (PersonId) REFERENCES User(Id)
    );


    CREATE TABLE IF NOT EXISTS Languages (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        PersonId BIGINT,
        Language VARCHAR(255),
        Level VARCHAR(255),
FOREIGN KEY (PersonId) REFERENCES User(Id)
    );

    CREATE TABLE IF NOT EXISTS Project (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
        PersonID BIGINT,
        Title VARCHAR(255),
        Role VARCHAR(255),
Description TEXT,
        Company VARCHAR(255),
        Document VARCHAR(255),
        ProjectStatus VARCHAR(255),
        StartDate DATE,
        ExpectedEndDate DATE,
IsOngoing BOOLEAN DEFAULT FALSE,
dmeta_title_primary VARCHAR(24) NOT NULL DEFAULT '',
dmeta_title_secondary VARCHAR(24) NOT NULL DEFAULT '',
FOREIGN KEY (PersonID) REFERENCES User(Id)
    );


-- Tabla de Notificaciones no de eventos
    CREATE TABLE IF NOT EXISTS Event (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
EventType VARCHAR(50) NOT NULL,
EventTitle VARCHAR(255) NOT NULL,
Description TEXT,
UserId BIGINT NOT NULL,
        OtherUserId BIGINT,
        ProyectId BIGINT,
CreateAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
IsRead BOOLEAN DEFAULT FALSE,
GroupId BIGINT,
Status VARCHAR(50) DEFAULT 'PENDING',
ActionRequired BOOLEAN DEFAULT FALSE,
ActionTakenAt DATETIME,
Metadata JSON,
dmeta_title_primary VARCHAR(24) NOT NULL DEFAULT '',
dmeta_title_secondary VARCHAR(24) NOT NULL DEFAULT '',
CreatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
UpdatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
FOREIGN KEY (UserId) REFERENCES User(Id),
FOREIGN KEY (OtherUserId) REFERENCES User(Id),
FOREIGN KEY (ProyectId) REFERENCES Project(Id),
FOREIGN KEY (GroupId) REFERENCES GroupsUsers(Id)
);



CREATE TABLE IF NOT EXISTS Notification (
Id BIGINT AUTO_INCREMENT PRIMARY KEY,
EventId BIGINT,
Description VARCHAR(255),
FOREIGN KEY (EventId) REFERENCES Event(Id)
);


CREATE TABLE IF NOT EXISTS CommunityEvent (
        Id BIGINT AUTO_INCREMENT PRIMARY KEY,
    -- Define qué tipo de publicación es, incluyendo 'DESAFIO'.
    PostType ENUM('EVENTO', 'NOTICIA', 'ARTICULO', 'ANUNCIO', 'MULTIMEDIA', 'DESAFIO', 'DISCUSION') NOT NULL DEFAULT 'EVENTO',

    Title VARCHAR(255) NOT NULL,
        Description TEXT,
    ImageUrl VARCHAR(255),

    -- Enlace principal (para artículos, noticias, videos o repositorios de desafíos)
    ContentUrl VARCHAR(2048) NULL,
    LinkPreviewTitle VARCHAR(255) NULL,
    LinkPreviewDescription VARCHAR(512) NULL,
    LinkPreviewImage VARCHAR(2048) NULL,

    -- Campos para EVENTOS
    EventDate DATETIME NULL,
        Location VARCHAR(255),
    Capacity INT NULL,
    Price DECIMAL(10, 2) NULL,
    
    -- --- NUEVOS CAMPOS PARA DESAFÍOS ---
    ChallengeStartDate DATETIME NULL,
    ChallengeEndDate DATETIME NULL,
    ChallengeDifficulty ENUM('PRINCIPIANTE', 'INTERMEDIO', 'AVANZADO', 'EXPERTO') NULL,
    ChallengePrize VARCHAR(512) NULL, -- Descripción del premio o recompensa
    ChallengeStatus ENUM('ABIERTO', 'EN_EVALUACION', 'CERRADO', 'CANCELADO') NOT NULL DEFAULT 'ABIERTO',

    -- --- CAMPOS COMUNES ---
    Tags JSON NULL, -- Puede usarse para tecnologías (ej: ['React', 'Node.js'])
    OrganizerCompanyName VARCHAR(255),
    OrganizerUserId BIGINT,
    OrganizerLogoUrl VARCHAR(255),
    CreatedByUserId BIGINT NOT NULL,
    dmeta_title_primary VARCHAR(24) NOT NULL DEFAULT '',
    dmeta_title_secondary VARCHAR(24) NOT NULL DEFAULT '',
    CreatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UpdatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (OrganizerUserId) REFERENCES User(Id) ON DELETE SET NULL,
    FOREIGN KEY (CreatedByUserId) REFERENCES User(Id) ON DELETE CASCADE
);



/*
Tabla ReputationReview
Descripción: Almacena cada evento de reseña y calificación entre dos usuarios de la plataforma.
Es el núcleo del sistema de reputación, registrando los Puntos de Reputación (RP) y
el feedback cualitativo.

Campos Principales:
- ReviewerId: El ID del usuario que realiza la calificación.
- RevieweeId: El ID del usuario que recibe la calificación.
- CommunityEventId: El ID del evento/publicación que origina la reseña. Esto es clave
  para permitir múltiples calificaciones entre los mismos usuarios pero en diferentes contextos.
- PointsRP: La cantidad de puntos crudos otorgados.
- Rating: La puntuación visible (ej. 4.5 estrellas).
- InteractionType: El contexto que originó la reseña.

Relaciones:
- Se vincula con la tabla User (dos veces) y con la tabla CommunityEvent.
*/
CREATE TABLE IF NOT EXISTS ReputationReview (
    Id BIGINT AUTO_INCREMENT PRIMARY KEY,

    -- Clave foránea que referencia al usuario que EMITE la reseña.
    ReviewerId BIGINT NOT NULL,

    -- Clave foránea que referencia al usuario que RECIBE la reseña y los puntos.
    RevieweeId BIGINT NOT NULL,

    -- --- CAMPO AÑADIDO ---
    -- Vincula la reseña a una publicación específica (oferta, evento, desafío).
    -- Es NOT NULL para asegurar que toda reseña tenga un contexto claro.
    CommunityEventId BIGINT NOT NULL,

    -- El valor numérico de "Puntos de Reputación" (RP).
    PointsRP INT NOT NULL,

    -- La calificación visible (ej. en una escala de 1 a 5).
    Rating DECIMAL(2, 1),

    -- El comentario o feedback cualitativo.
    Comment TEXT,

    -- Define el contexto de la reseña. Podría ser redundante con CommunityEvent.PostType
    -- pero se mantiene para flexibilidad.
    InteractionType ENUM('ENTREVISTA', 'MENTORIA', 'PROYECTO_COLABORATIVO', 'EVENTO', 'POSTULACION_EMPLEO', 'DESAFIO_COMPLETADO'),

    -- Timestamps
    CreatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Definición de las llaves foráneas.
    FOREIGN KEY (ReviewerId) REFERENCES User(Id) ON DELETE CASCADE,
    FOREIGN KEY (RevieweeId) REFERENCES User(Id) ON DELETE CASCADE,
    FOREIGN KEY (CommunityEventId) REFERENCES CommunityEvent(Id) ON DELETE CASCADE,

    -- --- NUEVA RESTRICCIÓN ---
    -- Asegura que solo pueda existir una única reseña por parte de un 'reviewer'
    -- a un 'reviewee' para un evento comunitario específico.
    UNIQUE KEY uq_unique_review_per_event (ReviewerId, RevieweeId, CommunityEventId)
);


CREATE TABLE IF NOT EXISTS FeedItemView (
    UserId BIGINT NOT NULL,
    -- ItemType distingue entre 'USER' (para perfiles de estudiante/empresa) y 'COMMUNITY_EVENT'
    ItemType ENUM('USER', 'COMMUNITY_EVENT') NOT NULL,
    ItemId BIGINT NOT NULL,
    ViewedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- Un usuario solo ve un item una vez. La PK previene duplicados.
    PRIMARY KEY (UserId, ItemType, ItemId),
    FOREIGN KEY (UserId) REFERENCES User(Id) ON DELETE CASCADE
);


CREATE TABLE IF NOT EXISTS JobApplication (
    Id BIGINT AUTO_INCREMENT PRIMARY KEY,

    -- --- La Conexión Clave ---
    -- Se conecta directamente con la publicación en la tabla CommunityEvent.
    CommunityEventId BIGINT NOT NULL,

    -- El usuario (estudiante/egresado) que se está postulando.
    ApplicantId BIGINT NOT NULL,

    -- El estado de la postulación dentro del proceso de selección.
    Status ENUM(
        'ENVIADA',
        'EN_REVISION',
        'ENTREVISTA',
        'PRUEBA_TECNICA',
        'OFERTA_REALIZADA',
        'APROBADA',
        'RECHAZADA',
        'RETIRADA'
    ) NOT NULL DEFAULT 'ENVIADA',

    -- Fecha en que se realizó la postulación.
    AppliedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Fecha de la última actualización del estado.
    UpdatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- Opcional: Un campo para una breve carta de presentación.
    CoverLetter TEXT,

    -- Definición de las llaves foráneas.
    FOREIGN KEY (CommunityEventId) REFERENCES CommunityEvent(Id) ON DELETE CASCADE,
    FOREIGN KEY (ApplicantId) REFERENCES User(Id) ON DELETE CASCADE,

    -- Restricción para asegurar que un usuario no pueda postularse dos veces a la misma oferta.
    UNIQUE KEY uq_event_applicant (CommunityEventId, ApplicantId)
    );
	`

	// Dividir el esquema en sentencias individuales
	statements := strings.Split(sqlSchema, ";")

	for _, stmt := range statements {
		trimmedStmt := strings.TrimSpace(stmt)
		if trimmedStmt == "" {
			continue // Saltar sentencias vacías resultantes del split
		}
		_, err := tx.Exec(trimmedStmt) // Ejecutar cada sentencia
		if err != nil {
			// Loguear la sentencia específica que falló para facilitar la depuración
			logger.Errorf("DB", "Error executing statement: %s", trimmedStmt)
			return fmt.Errorf("error executing schema creation statement: %w", err)
		}
	}

	logger.Success("DB", "Tables created or already exist.")
	return nil
}

// insertDefaultData populates tables with initial values, ignoring duplicates.
func insertDefaultData(tx *sql.Tx) error {
	logger.Info("DB", "Inserting default data...")

	// Insert Nationalities
	stmtNat, err := tx.Prepare("INSERT IGNORE INTO Nationality (CountryName, IsoCode, DocIdFormat) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare Nationality statement: %w", err)
	}
	defer stmtNat.Close()
	for _, nat := range models.GetDefaultNationalities() {
		_, err := stmtNat.Exec(nat.CountryName, nat.IsoCode, nat.DocIdFormat)
		if err != nil {
			logger.Warnf("DB", "Failed to insert nationality %s: %v", nat.CountryName, err)
			// Continue trying to insert others
		}
	}

	// Insert StatusAuthorized
	stmtStatus, err := tx.Prepare("INSERT IGNORE INTO StatusAuthorized (Id, Name) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare StatusAuthorized statement: %w", err)
	}
	defer stmtStatus.Close()
	for _, status := range models.GetDefaultStatusAuthorized() {
		_, err := stmtStatus.Exec(status.Id, status.Name)
		if err != nil {
			logger.Warnf("DB", "Failed to insert status %s: %v", status.Name, err)
		}
	}

	// Insert TokensType
	stmtToken, err := tx.Prepare("INSERT IGNORE INTO Token (Id, TokenType) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare Token statement: %w", err)
	}
	defer stmtToken.Close()
	for _, token := range models.GetDefaultTokensType() {
		_, err := stmtToken.Exec(token.Id, token.TokenType)
		if err != nil {
			logger.Warnf("DB", "Failed to insert token type %s: %v", token.TokenType, err)
		}
	}

	// Insert Roles
	stmtRole, err := tx.Prepare("INSERT IGNORE INTO Role (Id, Name) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare Role statement: %w", err)
	}
	defer stmtRole.Close()
	for _, role := range models.GetDefaultRoles() {
		_, err := stmtRole.Exec(role.Id, role.Name)
		if err != nil {
			logger.Warnf("DB", "Failed to insert role %s: %v", role.Name, err)
		}
	}

	// Insert Universities
	stmtUni, err := tx.Prepare("INSERT IGNORE INTO University (Name, Campus) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare University statement: %w", err)
	}
	defer stmtUni.Close()
	uniIDs := make(map[string]int64) // Store inserted university IDs
	for _, uni := range models.GetDefaultUniversities() {
		res, err := stmtUni.Exec(uni.Name, uni.Campus)
		if err != nil {
			logger.Warnf("DB", "Failed to insert university %s: %v", uni.Name, err)
			// Try to fetch existing ID if insertion failed due to duplicate
			row := tx.QueryRow("SELECT Id FROM University WHERE Name = ?", uni.Name)
			var id int64
			if scanErr := row.Scan(&id); scanErr == nil {
				uniIDs[uni.Name] = id
			} else {
				logger.Warnf("DB", "Failed to fetch existing ID for university %s: %v", uni.Name, scanErr)
			}
		} else {
			id, err := res.LastInsertId()
			if err == nil {
				uniIDs[uni.Name] = id
			} else {
				logger.Warnf("DB", "Failed to get last insert ID for university %s: %v", uni.Name, err)
			}
		}
	}

	// Insert Degrees (Requires University IDs)
	stmtDegree, err := tx.Prepare("INSERT IGNORE INTO Degree (DegreeName, Descriptions, Code, UniversityId) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare Degree statement: %w", err)
	}
	defer stmtDegree.Close()
	for _, deg := range models.GetDefaultDegrees() {
		// Find the corresponding University Name from the list to get the ID
		var uniName string
		for _, u := range models.GetDefaultUniversities() {
			if u.Id == deg.UniversityId { // Assuming original UniversityId in Degree refers to the implicit ID/index in the default list
				// THIS IS FRAGILE. It's better if Degree default data references University Name directly.
				// For now, we assume the default Degrees UniversityId = 1 maps to the first University in the list.
				if len(models.GetDefaultUniversities()) > 0 {
					uniName = models.GetDefaultUniversities()[0].Name
				}
				break
			}
		}

		uniID, ok := uniIDs[uniName]
		if !ok || uniName == "" {
			logger.Warnf("DB", "Skipping degree %s: Could not find University ID for assumed name '%s' (original ID: %d)", deg.DegreeName, uniName, deg.UniversityId)
			continue
		}

		_, err := stmtDegree.Exec(deg.DegreeName, deg.Descriptions, deg.Code, uniID)
		if err != nil {
			logger.Warnf("DB", "Failed to insert degree %s: %v", deg.DegreeName, err)
		}
	}

	// Insert TypeMessages
	stmtMsgType, err := tx.Prepare("INSERT IGNORE INTO TypeMessage (Id, Name, Description) VALUES (?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare TypeMessage statement: %w", err)
	}
	defer stmtMsgType.Close()
	for _, msgType := range models.GetDefaultTypeMessages() {
		_, err := stmtMsgType.Exec(msgType.Id, msgType.Name, msgType.Description)
		if err != nil {
			logger.Warnf("DB", "Failed to insert message type %s: %v", msgType.Name, err)
		}
	}

	logger.Success("DB", "Finished inserting default data.")
	return nil
}

// GetAllCategories retrieves all categories from the database, ordered by name.
func GetAllCategories() ([]models.Category, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	rows, err := db.Query("SELECT CategoryId, Name FROM Category ORDER BY Name ASC")
	if err != nil {
		logger.Errorf("DB", "Error querying categories: %v", err)
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.CategoryId, &category.Name); err != nil {
			logger.Errorf("DB", "Error scanning category row: %v", err)
			continue // Skip problematic rows
		}
		categories = append(categories, category)
	}
	if err = rows.Err(); err != nil {
		logger.Errorf("DB", "Error iterating category rows: %v", err)
		return nil, err
	}
	return categories, nil
}

// CheckCategoryExistsByName checks if a category with the given name already exists.
func CheckCategoryExistsByName(name string) (bool, error) {
	if db == nil {
		return false, fmt.Errorf("database not initialized")
	}
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM Category WHERE Name = ?)"
	err := db.QueryRow(query, name).Scan(&exists)
	if err != nil {
		logger.Errorf("DB", "Error checking category existence by name %s: %v", name, err)
		return false, err
	}
	return exists, nil
}

// AddCategory inserts a new category into the database.
// It assumes the name does not already exist (should be checked beforehand).
func AddCategory(name string) (models.Category, error) {
	if db == nil {
		return models.Category{}, fmt.Errorf("database not initialized")
	}
	query := "INSERT INTO Category (Name) VALUES (?)"
	result, err := db.Exec(query, name)
	if err != nil {
		logger.Errorf("DB", "Error inserting category %s: %v", name, err)
		// Consider checking for specific duplicate entry errors if needed
		return models.Category{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Errorf("DB", "Error getting last insert ID for category %s: %v", name, err)
		// The insert succeeded, but we can't get the ID easily.
		// Return a category with name only, or handle differently.
		return models.Category{Name: name}, err
	}

	return models.Category{CategoryId: id, Name: name}, nil
}
