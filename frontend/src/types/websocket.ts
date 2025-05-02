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