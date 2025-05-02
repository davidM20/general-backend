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
// ... otros tipos de mensajes necesarios ...

// Tipo para distinguir resultados en el frontend
export interface TypedUserSearchResult extends UserSearchResult {
  resultType: 'user';
}

export interface TypedEnterpriseSearchResult extends EnterpriseSearchResult {
  resultType: 'enterprise';
}

export type TypedSearchResult = TypedUserSearchResult | TypedEnterpriseSearchResult; 