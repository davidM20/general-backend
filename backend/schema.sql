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

-- Índice para búsquedas por nombre y apellido (búsquedas de personas)
CREATE INDEX idx_user_name ON User(FirstName, LastName);

-- Índice para filtrar usuarios por rol y estado (muy común en consultas de administración)
CREATE INDEX idx_user_role_status ON User(RoleId, StatusAuthorizedId);

-- Índice para búsquedas académicas (filtrar por universidad/carrera)
CREATE INDEX idx_user_academic ON User(UniversityId, DegreeId);

-- Índice para ordenamiento por fecha de creación (para listados recientes)
CREATE INDEX idx_user_created ON User(CreatedAt);

-- Índice para búsquedas de empresas por sector
CREATE INDEX idx_company_sector ON User(Sector, RoleId);

-- Índice para búsquedas por ubicación
CREATE INDEX idx_user_location ON User(Location);

-- Índice para búsquedas por RIF
CREATE INDEX idx_user_rif ON User(RIF);

-- Índice para búsquedas por CompanyName
CREATE INDEX idx_user_company_name ON User(CompanyName);

-- Columnas para búsqueda fonética de nombres de personas y empresas
CREATE INDEX idx_user_phonetic_person ON User(dmeta_person_primary, dmeta_person_secondary);
CREATE INDEX idx_user_phonetic_company ON User(dmeta_company_primary, dmeta_company_secondary);



-- Modificaciones adicionales a la tabla User
-- Agregar campos de empresa
ALTER TABLE User
ADD COLUMN RIF VARCHAR(20) UNIQUE,
ADD COLUMN Sector VARCHAR(100),
ADD COLUMN CompanyName VARCHAR(255),
ADD COLUMN Location VARCHAR(255),
ADD COLUMN FoundationYear INT,
ADD COLUMN EmployeeCount INT,
ADD COLUMN CreatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN UpdatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;

-- Agregar campos de redes sociales y contacto
ALTER TABLE User
ADD COLUMN ContactEmail VARCHAR(255),
ADD COLUMN Twitter VARCHAR(255),
ADD COLUMN Facebook VARCHAR(255);

-- Agregar índices para mejorar el rendimiento
ALTER TABLE User
ADD INDEX idx_rif (RIF),
ADD INDEX idx_company_name (CompanyName),
ADD INDEX idx_sector (Sector);

-- Agregar restricciones
ALTER TABLE User
ADD CONSTRAINT chk_employee_count CHECK (EmployeeCount >= 0),
ADD CONSTRAINT chk_foundation_year CHECK (FoundationYear > 1800 AND FoundationYear <= 2100);

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
CreateAt DATE,
ContentId VARCHAR(255),
ChatId VARCHAR(255)
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
- Semántica de nombres: Se han renombrado campos como `UserId` a `SenderId` y `ResponseTo` a `ReplyToMessageId` para mayor claridad.
- Contenido del mensaje: `Text` se cambia a `Content` y su tipo a `TEXT` para permitir mensajes más largos.
- Timestamps precisos: `Date` (que solo guardaba la fecha) se reemplaza por `SentAt` (DATETIME) para incluir la hora y se añade `EditedAt` para registrar ediciones.
- Estado del mensaje: `StatusMessage` (INT) se convierte en un `ENUM` para que los valores sean auto-descriptivos ('sending', 'sent', 'delivered', 'read', 'failed').
- Integridad de datos: Se añaden restricciones (CHECK constraints) para:
    1. Asegurar que un mensaje pertenezca a un chat (`ChatId`) O a un grupo (`ChatIdGroup`), pero no a ambos.
    2. Evitar mensajes vacíos (debe tener `Content` o `MediaId`).
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

-- Índices para la tabla Message
-- Optimiza la búsqueda de mensajes dentro de un chat privado, ordenados por fecha.
-- Es la consulta más común al abrir una conversación.
CREATE INDEX idx_message_chat_sent ON Message(ChatId, SentAt DESC);

-- Optimiza la búsqueda de mensajes dentro de un chat de grupo, ordenados por fecha.
CREATE INDEX idx_message_group_sent ON Message(ChatIdGroup, SentAt DESC);

-- Acelera la búsqueda de todos los mensajes enviados por un usuario.
CREATE INDEX idx_message_sender ON Message(SenderId);

-- Optimiza el conteo de mensajes no leídos para un usuario en un chat.
-- Nota: para contar no leídos para un usuario específico, necesitarías incluir SenderId != current_user_id en tu query.
CREATE INDEX idx_message_chat_status ON Message(ChatId, Status);

CREATE INDEX idx_message_group_status ON Message(ChatIdGroup, Status);

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

CREATE INDEX idx_education_phonetic_institution ON Education(dmeta_institution_primary, dmeta_institution_secondary);
CREATE INDEX idx_education_phonetic_degree ON Education(dmeta_degree_primary, dmeta_degree_secondary);

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

CREATE INDEX idx_workexperience_phonetic_company ON WorkExperience(dmeta_company_primary, dmeta_company_secondary);
CREATE INDEX idx_workexperience_phonetic_position ON WorkExperience(dmeta_position_primary, dmeta_position_secondary);

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

CREATE INDEX idx_skill_phonetic ON Skills(dmeta_primary, dmeta_secondary);

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

CREATE INDEX idx_project_phonetic_title ON Project(dmeta_title_primary, dmeta_title_secondary);

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

   CREATE INDEX idx_event_user_status ON Event(UserId, Status);
   CREATE INDEX idx_event_user_isread ON Event(UserId, IsRead);

   CREATE INDEX idx_event_createat ON Event(CreateAt);
   CREATE INDEX idx_event_actiontakenat ON Event(ActionTakenAt);

   CREATE INDEX idx_event_type_status ON Event(EventType, Status);
   CREATE INDEX idx_event_actionrequired_isread ON Event(ActionRequired, IsRead);

   CREATE INDEX idx_event_project ON Event(ProyectId);
   CREATE INDEX idx_event_group ON Event(GroupId);
   CREATE INDEX idx_event_otheruserid ON Event(OtherUserId);

   CREATE INDEX idx_community_event_date ON CommunityEvent(EventDate);
   CREATE INDEX idx_community_event_created_at ON CommunityEvent(CreatedAt);
   CREATE INDEX idx_community_event_organizer_user ON CommunityEvent(OrganizerUserId);
   CREATE INDEX idx_community_event_created_by ON CommunityEvent(CreatedByUserId);
   CREATE INDEX idx_community_event_phonetic_title ON CommunityEvent(dmeta_title_primary, dmeta_title_secondary);


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


-- Índices para CommunityEvent
CREATE INDEX idx_community_event_date ON CommunityEvent(EventDate);
CREATE INDEX idx_community_event_created_at ON CommunityEvent(CreatedAt);
CREATE INDEX idx_community_event_organizer_user ON CommunityEvent(OrganizerUserId);
CREATE INDEX idx_community_event_created_by ON CommunityEvent(CreatedByUserId);
CREATE INDEX idx_community_event_phonetic_title ON CommunityEvent(dmeta_title_primary, dmeta_title_secondary);

-- Comandos ALTER TABLE para aplicar estos cambios a una tabla existente:
ALTER TABLE CommunityEvent ADD COLUMN Capacity INT NULL;
ALTER TABLE CommunityEvent ADD COLUMN Price DECIMAL(10, 2) NULL;
ALTER TABLE CommunityEvent ADD COLUMN Tags JSON NULL;

-- ALTER commands for existing databases to apply new schema changes
ALTER TABLE User MODIFY Summary TEXT;
ALTER TABLE WorkExperience MODIFY Description TEXT;
ALTER TABLE Project MODIFY Description TEXT;
ALTER TABLE Education ADD COLUMN IsCurrentlyStudying BOOLEAN DEFAULT FALSE;
ALTER TABLE WorkExperience ADD COLUMN IsCurrentJob BOOLEAN DEFAULT FALSE;
ALTER TABLE Project ADD COLUMN IsOngoing BOOLEAN DEFAULT FALSE;




-- Acelera la búsqueda de contactos por el primer usuario y su estado.
CREATE INDEX idx_contact_user1_status ON Contact (User1Id, Status);

-- Acelera la búsqueda de contactos por el segundo usuario y su estado.
CREATE INDEX idx_contact_user2_status ON Contact (User2Id, Status);

-- Acelera las uniones (JOINs) con los mensajes usando el ChatId.
CREATE INDEX idx_contact_chatid ON Contact (ChatId);


-- Se han movido y mejorado los índices de la tabla Message justo después de su definición.


-- Comandos ALTER para migrar la tabla Message existente a la nueva estructura.
-- Ejecutar en orden.
-- 1. Renombrar columnas y modificar tipos
ALTER TABLE Message CHANGE COLUMN UserId SenderId BIGINT NOT NULL;
ALTER TABLE Message CHANGE COLUMN Text Content TEXT;
ALTER TABLE Message CHANGE COLUMN ResponseTo ReplyToMessageId VARCHAR(255);
ALTER TABLE Message CHANGE COLUMN `Date` SentAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- 2. Añadir la nueva columna de estado y eliminar la antigua
ALTER TABLE Message ADD COLUMN Status ENUM('sending', 'sent', 'delivered', 'read', 'failed') NOT NULL DEFAULT 'sending';
-- Opcional: Migrar datos del antiguo StatusMessage (INT) al nuevo Status (ENUM)
-- UPDATE Message SET Status = CASE StatusMessage WHEN 1 THEN 'sent' WHEN 2 THEN 'delivered' WHEN 3 THEN 'read' ELSE 'sent' END;
ALTER TABLE Message DROP COLUMN StatusMessage;

-- 3. Añadir la columna para la fecha de edición
ALTER TABLE Message ADD COLUMN EditedAt DATETIME;

-- 4. Añadir las restricciones de integridad
ALTER TABLE Message ADD CONSTRAINT chk_message_content CHECK (Content IS NOT NULL OR MediaId IS NOT NULL);
ALTER TABLE Message ADD CONSTRAINT chk_message_chat_or_group CHECK (
    (ChatId IS NOT NULL AND ChatIdGroup IS NULL) OR
    (ChatId IS NULL AND ChatIdGroup IS NOT NULL)
);

-- 5. Eliminar los índices antiguos (si existen)
DROP INDEX IF EXISTS idx_message_chatid_date_id ON Message;
DROP INDEX IF EXISTS idx_message_chatid_userid_status ON Message;

-- 6. Crear los nuevos índices (ya deberían estar en la definición de la tabla nueva)
-- CREATE INDEX idx_message_chat_sent ON Message(ChatId, SentAt DESC);
-- CREATE INDEX idx_message_group_sent ON Message(ChatIdGroup, SentAt DESC);
-- CREATE INDEX idx_message_sender ON Message(SenderId);
-- CREATE INDEX idx_message_chat_status ON Message(ChatId, Status);
-- CREATE INDEX idx_message_group_status ON Message(ChatIdGroup, Status);



/*
Tabla ReputationReview
Descripción: Almacena cada evento de reseña y calificación entre dos usuarios de la plataforma.
Es el núcleo del sistema de reputación, registrando los Puntos de Reputación (RP) y
el feedback cualitativo.

Campos Principales:
- IdEvaluador: El ID del usuario (persona o empresa) que realiza la calificación.
- IdEvaluado: El ID del usuario (persona o empresa) que recibe la calificación y los puntos.
- PuntosRP: La cantidad de puntos crudos otorgados, usados para el cálculo del nivel logarítmico.
- Calificacion: La puntuación visible (ej. 4.5 estrellas).
- TipoInteraccion: El contexto que originó la reseña (ej. una entrevista).

Relaciones:
- Se vincula dos veces con la tabla User a través de IdEvaluador y IdEvaluado.
*/

CREATE TABLE IF NOT EXISTS ReputationReview (
    Id BIGINT AUTO_INCREMENT PRIMARY KEY,

    -- Clave foránea que referencia al usuario que EMITE la reseña.
    -- Puede ser un estudiante, egresado o una empresa.
    ReviewerId BIGINT NOT NULL,

    -- Clave foránea que referencia al usuario que RECIBE la reseña y los puntos.
    -- Puede ser un estudiante, egresado o una empresa.
    RevieweeId BIGINT NOT NULL,

    -- El valor numérico de "Puntos de Reputación" (RP) otorgados en esta interacción.
    -- Este es el valor que se suma para el sistema de niveles logarítmico.
    PointsRP INT NOT NULL,

    -- La calificación visible por los usuarios (ej. una escala de 1 a 5).
    Rating DECIMAL(2, 1),

    -- El comentario o feedback cualitativo.
    Comment TEXT,

    -- Define el contexto de la reseña para dar más información.
    InteractionType ENUM('ENTREVISTA', 'MENTORIA', 'PROYECTO_COLABORATIVO', 'EVENTO'),

    -- Timestamps
    CreatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Definición de las llaves foráneas, vinculando esta tabla con la tabla User.
    FOREIGN KEY (ReviewerId) REFERENCES User(Id),
    FOREIGN KEY (RevieweeId) REFERENCES User(Id)
);

-- =================================================================
-- MIGRACIÓN PARA BÚSQUEDA FONÉTICA (DOUBLE METAPHONE)
-- =================================================================

-- Añadir columnas fonéticas e índices a la tabla User
ALTER TABLE User
ADD COLUMN dmeta_person_primary VARCHAR(24) NOT NULL DEFAULT '' AFTER EmployeeCount,
ADD COLUMN dmeta_person_secondary VARCHAR(24) NOT NULL DEFAULT '' AFTER dmeta_person_primary,
ADD COLUMN dmeta_company_primary VARCHAR(24) NOT NULL DEFAULT '' AFTER dmeta_person_secondary,
ADD COLUMN dmeta_company_secondary VARCHAR(24) NOT NULL DEFAULT '' AFTER dmeta_company_primary;

ALTER TABLE User
ADD INDEX idx_user_phonetic_person (dmeta_person_primary, dmeta_person_secondary),
ADD INDEX idx_user_phonetic_company (dmeta_company_primary, dmeta_company_secondary);

-- Añadir columnas fonéticas e índices a la tabla Education
ALTER TABLE Education
ADD COLUMN dmeta_institution_primary VARCHAR(24) NOT NULL DEFAULT '' AFTER IsCurrentlyStudying,
ADD COLUMN dmeta_institution_secondary VARCHAR(24) NOT NULL DEFAULT '' AFTER dmeta_institution_primary,
ADD COLUMN dmeta_degree_primary VARCHAR(24) NOT NULL DEFAULT '' AFTER dmeta_institution_secondary,
ADD COLUMN dmeta_degree_secondary VARCHAR(24) NOT NULL DEFAULT '' AFTER dmeta_degree_primary;

ALTER TABLE Education
ADD INDEX idx_education_phonetic_institution (dmeta_institution_primary, dmeta_institution_secondary),
ADD INDEX idx_education_phonetic_degree (dmeta_degree_primary, dmeta_degree_secondary);

-- Añadir columnas fonéticas e índices a la tabla WorkExperience
ALTER TABLE WorkExperience
ADD COLUMN dmeta_company_primary VARCHAR(24) NOT NULL DEFAULT '' AFTER IsCurrentJob,
ADD COLUMN dmeta_company_secondary VARCHAR(24) NOT NULL DEFAULT '' AFTER dmeta_company_primary,
ADD COLUMN dmeta_position_primary VARCHAR(24) NOT NULL DEFAULT '' AFTER dmeta_company_secondary,
ADD COLUMN dmeta_position_secondary VARCHAR(24) NOT NULL DEFAULT '' AFTER dmeta_position_primary;

ALTER TABLE WorkExperience
ADD INDEX idx_workexperience_phonetic_company (dmeta_company_primary, dmeta_company_secondary),
ADD INDEX idx_workexperience_phonetic_position (dmeta_position_primary, dmeta_position_secondary);

-- Añadir columnas fonéticas e índices a la tabla Skills
ALTER TABLE Skills
ADD COLUMN dmeta_primary VARCHAR(12) NOT NULL DEFAULT '' AFTER Level,
ADD COLUMN dmeta_secondary VARCHAR(12) NOT NULL DEFAULT '' AFTER dmeta_primary;

ALTER TABLE Skills
ADD INDEX idx_skill_phonetic (dmeta_primary, dmeta_secondary);

-- Añadir columnas fonéticas e índices a la tabla Project
ALTER TABLE Project
ADD COLUMN dmeta_title_primary VARCHAR(24) NOT NULL DEFAULT '' AFTER IsOngoing,
ADD COLUMN dmeta_title_secondary VARCHAR(24) NOT NULL DEFAULT '' AFTER dmeta_title_primary;

ALTER TABLE Project
ADD INDEX idx_project_phonetic_title (dmeta_title_primary, dmeta_title_secondary);

-- Añadir columnas fonéticas e índices a la tabla CommunityEvent
ALTER TABLE CommunityEvent
ADD COLUMN dmeta_title_primary VARCHAR(24) NOT NULL DEFAULT '' AFTER CreatedByUserId,
ADD COLUMN dmeta_title_secondary VARCHAR(24) NOT NULL DEFAULT '' AFTER dmeta_title_primary;

ALTER TABLE CommunityEvent
ADD INDEX idx_community_event_phonetic_title (dmeta_title_primary, dmeta_title_secondary);


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

-- Índice para permitir la limpieza eficiente de registros de vistas antiguos si es necesario.
CREATE INDEX idx_feeditemview_viewedat ON FeedItemView(ViewedAt);


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

-- Índices para consultas de alto rendimiento:
-- Para que una empresa pueda ver rápidamente todos los postulantes a su oferta.
CREATE INDEX idx_jobapplication_event_status ON JobApplication(CommunityEventId, Status);
-- Para que un usuario pueda ver el estado de todas sus postulaciones.
CREATE INDEX idx_jobapplication_applicant_status ON JobApplication(ApplicantId, Status);