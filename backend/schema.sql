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
DegreeId BIGINT,
UniversityId BIGINT,
RoleId INT,
StatusAuthorizedId INT,
Summary VARCHAR(255),
Address VARCHAR(255),
Github VARCHAR(255),
Linkedin VARCHAR(255),
RIF VARCHAR(20) UNIQUE,
Sector VARCHAR(100),
CompanyName VARCHAR(255),
Location VARCHAR(255),
FoundationYear INT,
EmployeeCount INT,
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
Status VARCHAR(255),
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

CREATE TABLE IF NOT EXISTS Message (
Id VARCHAR(255) PRIMARY KEY,
TypeMessageId BIGINT,
Text VARCHAR(255),
MediaId VARCHAR(255),
Date DATE,
StatusMessage INT,
UserId BIGINT,
ChatId VARCHAR(255),
ChatIdGroup VARCHAR(255),
ResponseTo VARCHAR(255),
FOREIGN KEY (TypeMessageId) REFERENCES TypeMessage(Id),
FOREIGN KEY (MediaId) REFERENCES Multimedia(Id),
FOREIGN KEY (UserId) REFERENCES User(Id),
FOREIGN KEY (ChatId) REFERENCES Contact(ChatId),
FOREIGN KEY (ChatIdGroup) REFERENCES GroupsUsers(ChatId),
FOREIGN KEY (ResponseTo) REFERENCES Message(Id)
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
FOREIGN KEY (PersonId) REFERENCES User(Id)
);

CREATE TABLE IF NOT EXISTS WorkExperience (
Id BIGINT AUTO_INCREMENT PRIMARY KEY,
PersonId BIGINT,
Company VARCHAR(255),
Position VARCHAR(255),
StartDate DATE,
EndDate DATE,
Description VARCHAR(255),
CountryId BIGINT,
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
Description VARCHAR(255),
Company VARCHAR(255),
Document VARCHAR(255),
ProjectStatus VARCHAR(255),
StartDate DATE,
ExpectedEndDate DATE,
FOREIGN KEY (PersonID) REFERENCES User(Id)
);

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


CREATE TABLE IF NOT EXISTS Notification (
Id BIGINT AUTO_INCREMENT PRIMARY KEY,
EventId BIGINT,
Description VARCHAR(255),
FOREIGN KEY (EventId) REFERENCES Event(Id)
);
