-- Primero, verificar la estructura actual de la tabla
DESCRIBE Event;

-- Modificación de la tabla Event para soportar notificaciones mejoradas

-- 1. Eliminar las restricciones de clave foránea existentes
ALTER TABLE Event
    DROP FOREIGN KEY Event_ibfk_1,
    DROP FOREIGN KEY Event_ibfk_2,
    DROP FOREIGN KEY Event_ibfk_3,
    DROP FOREIGN KEY Event_ibfk_4;

-- 2. Realizar las modificaciones de la tabla
ALTER TABLE Event
    ADD COLUMN EventType VARCHAR(50) NOT NULL AFTER Id,
    ADD COLUMN EventTitle VARCHAR(255) NOT NULL AFTER EventType,
    MODIFY COLUMN Description TEXT,
    MODIFY COLUMN UserId BIGINT NOT NULL,
    MODIFY COLUMN CreateAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN Status VARCHAR(50) DEFAULT 'PENDING',
    ADD COLUMN ActionRequired BOOLEAN DEFAULT FALSE,
    ADD COLUMN ActionTakenAt DATETIME,
    ADD COLUMN Metadata JSON;

-- 3. Volver a crear las restricciones de clave foránea
ALTER TABLE Event
    ADD CONSTRAINT Event_ibfk_1 FOREIGN KEY (UserId) REFERENCES User(Id),
    ADD CONSTRAINT Event_ibfk_2 FOREIGN KEY (OtherUserId) REFERENCES User(Id),
    ADD CONSTRAINT Event_ibfk_3 FOREIGN KEY (ProyectId) REFERENCES Project(Id),
    ADD CONSTRAINT Event_ibfk_4 FOREIGN KEY (GroupId) REFERENCES GroupsUsers(Id);

-- Comentarios sobre los cambios
-- EventType: Nuevo campo para el tipo de evento
-- EventTitle: Nuevo campo para el título del evento
-- Description: Cambiado a TEXT para permitir descripciones más largas
-- UserId: Ahora es NOT NULL
-- CreateAt: Ahora tiene valor por defecto
-- Status: Nuevo campo para manejar el estado de la notificación
-- ActionRequired: Indica si la notificación requiere acción
-- ActionTakenAt: Registra cuándo se tomó la acción
-- Metadata: Campo JSON para datos adicionales específicos 