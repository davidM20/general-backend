// src/types/websocket.ts

// (Añadir al contenido existente o crear el archivo)

// --- Tipos para Búsqueda ---

export interface SearchPayload {
  query: string;
  entityType: 'user' | 'enterprise' | 'all';
  limit?: number;
  offset?: number;
}

export interface UserSearchResult {
  id: number; // Go usa int64, TS usa number
  firstName?: string;
  lastName?: string;
  userName: string;
  email: string;
  picture?: string;
  summary?: string;
  roleId: number;
  roleName?: string;
  universityId?: number;
  universityName?: string;
  degreeId?: number;
  degreeName?: string;
}

export interface EnterpriseSearchResult {
  id: number;
  rif: string;
  companyName: string;
  categoryId?: number;
  categoryName?: string;
  description?: string;
  location?: string;
  phone?: string;
}

export interface SearchResponsePayload {
  query: string;
  entityType: 'user' | 'enterprise' | 'all';
  results: (UserSearchResult | EnterpriseSearchResult)[]; // Array de resultados mixtos
}

// Interfaz para el mensaje saliente genérico si aún no existe
export interface OutgoingMessage<T = any> {
  type: string;
  payload: T;
}

// Interfaz para el mensaje entrante genérico si aún no existe
export interface IncomingMessage<T = any> {
  type: string;
  payload: T;
  error?: string;
}

// Definir los tipos de mensajes si aún no existen
export const MessageTypeSearch = "search";
export const MessageTypeSearchResponse = "search_response";
export const MessageTypeError = "error";
export const MessageTypeGetMyProfile = "get_my_profile";
export const MessageTypeGetProfileResponse = "get_profile_response";
// ... otros tipos de mensajes necesarios ...

// Tipo para distinguir resultados en el frontend
export interface TypedUserSearchResult extends UserSearchResult {
  resultType: 'user';
}

export interface TypedEnterpriseSearchResult extends EnterpriseSearchResult {
  resultType: 'enterprise';
}

export type TypedSearchResult = TypedUserSearchResult | TypedEnterpriseSearchResult;

// --- Tipos para Perfil y Currículum ---

// Definir tipos para el Currículum si no existen
export interface EducationResponse {
  id: number;
  institution: string;
  degree: string;
  campus: string;
  graduation_date?: string; // Fecha como string ISO 8601
  country_id: number;
}

export interface WorkExperienceResponse {
  id: number;
  company: string;
  position: string;
  start_date?: string;
  end_date?: string;
  description: string;
  country_id: number;
}

export interface CertificationsResponse {
  id: number;
  certification: string;
  institution: string;
  date_obtained?: string;
}

export interface SkillsResponse {
  id: number;
  person_id?: number; // Suele omitirse en respuesta
  skill: string;
  level: string;
}

export interface LanguagesResponse {
  id: number;
  person_id?: number;
  language: string;
  level: string;
}

export interface ProjectResponse {
  id: number;
  person_id?: number;
  title: string;
  role: string;
  description: string;
  company: string;
  document: string;
  project_status: string;
  start_date?: string;
  expected_end_date?: string;
}

export interface CurriculumResponse {
  education?: EducationResponse[];
  experience?: WorkExperienceResponse[];
  skills?: SkillsResponse[];
  certifications?: CertificationsResponse[];
  languages?: LanguagesResponse[];
  project?: ProjectResponse[]; // Asegúrate que el nombre coincida (Project vs project)
}

export interface MyProfileResponse {
  id: number;
  firstName?: string;
  lastName?: string;
  userName: string;
  email: string;
  phone?: string;
  sex?: string;
  docId?: string;
  nationalityId?: number;
  nationalityName?: string;
  birthdate?: string; // Fecha como string ISO 8601 Z
  picture?: string;
  degreeId?: number;
  degreeName?: string;
  universityId?: number;
  universityName?: string;
  roleId: number;
  roleName?: string;
  statusAuthorizedId: number;
  summary?: string;
  address?: string;
  github?: string;
  linkedin?: string;
  curriculum: CurriculumResponse; // Usar el tipo de currículum
}

// --- Tipos para Listas --- 

// Payload para solicitar una lista (ya debería existir)
// export interface ListRequestPayload { listType: string; ... }

// Tipos de Info para las listas
export interface ContactInfo {
  userId: number;
  userName: string;
  firstName?: string;
  lastName?: string;
  picture?: string;
  isOnline: boolean;
  chatId: string;
}

export interface ChatInfo {
  chatId: string;
  otherUserId: number;
  otherUserName: string;
  otherFirstName?: string;
  otherLastName?: string;
  otherPicture?: string;
  lastMessage?: string;
  lastMessageTs?: number; // timestamp ms
  unreadCount?: number;
  isOnline: boolean; // isOtherOnline en Go, renombrado por consistencia?
}

export interface OnlineUserInfo {
  userId: number;
  userName: string;
}

// Tipos de respuesta para las listas (añadir si faltan)
export const MessageTypeListContactsResponse = "list_contacts_response";
export const MessageTypeListChatsResponse = "list_chats_response";
export const MessageTypeListOnlineUsersResponse = "list_online_users_response";

// Payload genérico para respuestas de listas (ya debería existir)
// export interface ListResponsePayload { listType: string; data: any; }

// Payloads específicos para las respuestas de lista
export interface ListContactsResponsePayload {
    data: ContactInfo[];
}
export interface ListChatsResponsePayload {
    data: ChatInfo[];
}
export interface ListOnlineUsersResponsePayload {
    data: OnlineUserInfo[];
}

// --- Tipos para Notificaciones ---

// Tipo para solicitar notificaciones
export const MessageTypeGetNotifications = "get-notifications"

// Tipo y Payload para la respuesta de notificaciones
export const MessageTypeNotificationsResponse = "notifications_response"; // CONFIRMAR este tipo con el backend

export interface NotificationInfo {
  id: number; // o string si es UUID
  // Ajustar campos según la tabla Event o lo que envíe el handler
  description: string;
  otherUserId?: number;
  otherUserName?: string; // Puede ser útil añadirlo
  projectId?: number;
  projectTitle?: string; // Puede ser útil
  createdAt: string; // Fecha como string ISO 8601 Z
  isRead?: boolean; // Si el backend marca el estado
  // Otros campos relevantes...
}

export interface NotificationsResponsePayload {
    notifications: NotificationInfo[];
} 