# Guía Completa: WebSocket Client en React Native con Expo

## Tabla de Contenidos
1. [Instalación y Configuración](#instalación-y-configuración)
2. [Arquitectura Recomendada](#arquitectura-recomendada)
3. [Context Provider](#context-provider)
4. [Hooks Personalizados](#hooks-personalizados)
5. [Implementación en Pantallas](#implementación-en-pantallas)
6. [Ejemplos Prácticos](#ejemplos-prácticos)
7. [Mejores Prácticas](#mejores-prácticas)
8. [Troubleshooting](#troubleshooting)

## Instalación y Configuración

### Prerequisitos
```bash
# Instalar dependencias necesarias
npx expo install react-native-url-polyfill
```

### Configuración inicial en App.js
```javascript
// App.js
import 'react-native-url-polyfill/auto'; // Necesario para el polyfill de URL en React Native
import { WebSocketProvider } from './contexts/WebSocketContext';

export default function App() {
  return (
    <WebSocketProvider>
      {/* Tu aplicación aquí */}
      <Navigation />
    </WebSocketProvider>
  );
}
```

## Arquitectura Recomendada

La arquitectura más eficiente para múltiples pantallas utiliza:
- **Context API** para compartir la instancia del WebSocket
- **Hooks personalizados** para funcionalidades específicas
- **Estado global** para datos compartidos
- **Callbacks centralizados** para manejo de eventos

```
App
├── WebSocketProvider (Context)
├── Pantalla 1 (useWebSocket, useChatMessages)
├── Pantalla 2 (useWebSocket, usePresence)
└── Pantalla 3 (useWebSocket, useJobSearch)
```

## Context Provider

Crea el archivo `contexts/WebSocketContext.js`:

```javascript
// contexts/WebSocketContext.js
import React, { createContext, useContext, useReducer, useRef, useCallback, useEffect } from 'react';
import { Alert } from 'react-native';
import { createWebSocketClient } from '../websocket-client'; // Tu archivo .ts convertido a .js

// Estados del WebSocket
const initialState = {
  connectionState: 'disconnected',
  isConnected: false,
  chatMessages: [],
  presenceUsers: {},
  jobSearchResults: [],
  notifications: [],
  userProfile: null,
  pendingAcks: 0,
  pendingResponses: 0
};

// Tipos de acciones
const actionTypes = {
  SET_CONNECTION_STATE: 'SET_CONNECTION_STATE',
  SET_USER_PROFILE: 'SET_USER_PROFILE',
  ADD_CHAT_MESSAGE: 'ADD_CHAT_MESSAGE',
  UPDATE_PRESENCE: 'UPDATE_PRESENCE',
  SET_JOB_SEARCH_RESULTS: 'SET_JOB_SEARCH_RESULTS',
  ADD_NOTIFICATION: 'ADD_NOTIFICATION',
  UPDATE_PENDING_COUNTS: 'UPDATE_PENDING_COUNTS',
  CLEAR_NOTIFICATIONS: 'CLEAR_NOTIFICATIONS'
};

// Reducer
function webSocketReducer(state, action) {
  switch (action.type) {
    case actionTypes.SET_CONNECTION_STATE:
      return {
        ...state,
        connectionState: action.payload,
        isConnected: action.payload === 'connected'
      };
    
    case actionTypes.SET_USER_PROFILE:
      return {
        ...state,
        userProfile: action.payload
      };
    
    case actionTypes.ADD_CHAT_MESSAGE:
      return {
        ...state,
        chatMessages: [...state.chatMessages, action.payload]
      };
    
    case actionTypes.UPDATE_PRESENCE:
      return {
        ...state,
        presenceUsers: {
          ...state.presenceUsers,
          [action.payload.userId]: action.payload.status
        }
      };
    
    case actionTypes.SET_JOB_SEARCH_RESULTS:
      return {
        ...state,
        jobSearchResults: action.payload
      };
    
    case actionTypes.ADD_NOTIFICATION:
      return {
        ...state,
        notifications: [...state.notifications, action.payload]
      };
    
    case actionTypes.UPDATE_PENDING_COUNTS:
      return {
        ...state,
        pendingAcks: action.payload.acks,
        pendingResponses: action.payload.responses
      };
    
    case actionTypes.CLEAR_NOTIFICATIONS:
      return {
        ...state,
        notifications: []
      };
    
    default:
      return state;
  }
}

// Context
const WebSocketContext = createContext();

// Provider Component
export function WebSocketProvider({ children }) {
  const [state, dispatch] = useReducer(webSocketReducer, initialState);
  const wsClient = useRef(null);
  const isInitialized = useRef(false);

  // Inicializar WebSocket
  const initializeWebSocket = useCallback(async (config) => {
    if (isInitialized.current) return;

    try {
      wsClient.current = createWebSocketClient(
        {
          url: config.url,
          authParams: new URLSearchParams({ token: config.token }),
          enableDebugLogs: __DEV__, // Solo en desarrollo
          reconnectInterval: 3000,
          maxReconnectAttempts: 5,
          ...config
        },
        {
          // Callbacks del WebSocket
          onConnect: (userData) => {
            console.log('WebSocket conectado:', userData);
            dispatch({ type: actionTypes.SET_USER_PROFILE, payload: userData });
            updatePendingCounts();
          },
          
          onDisconnect: (error) => {
            console.log('WebSocket desconectado:', error);
            if (error) {
              dispatch({
                type: actionTypes.ADD_NOTIFICATION,
                payload: {
                  id: Date.now(),
                  type: 'error',
                  message: 'Conexión perdida. Reintentando...',
                  timestamp: new Date().toISOString()
                }
              });
            }
          },
          
          onConnectionStateChange: (connectionState) => {
            dispatch({ type: actionTypes.SET_CONNECTION_STATE, payload: connectionState });
          },
          
          onChatEvent: (message) => {
            dispatch({
              type: actionTypes.ADD_CHAT_MESSAGE,
              payload: {
                id: message.pid,
                fromUserId: message.fromUserId,
                text: message.payload.text,
                timestamp: new Date().toISOString(),
                type: 'received'
              }
            });
          },
          
          onPresenceEvent: (message) => {
            dispatch({
              type: actionTypes.UPDATE_PRESENCE,
              payload: {
                userId: message.fromUserId,
                status: message.payload.status
              }
            });
          },
          
          onJobSearchResult: (message) => {
            dispatch({
              type: actionTypes.SET_JOB_SEARCH_RESULTS,
              payload: message.payload.results || []
            });
          },
          
          onErrorNotification: (message) => {
            Alert.alert(
              'Error',
              message.error?.message || 'Error desconocido',
              [{ text: 'OK' }]
            );
          }
        }
      );

      await wsClient.current.connect();
      isInitialized.current = true;
      
    } catch (error) {
      console.error('Error inicializando WebSocket:', error);
      Alert.alert('Error de Conexión', 'No se pudo conectar al servidor');
    }
  }, []);

  // Actualizar contadores de pendientes
  const updatePendingCounts = useCallback(() => {
    if (wsClient.current) {
      dispatch({
        type: actionTypes.UPDATE_PENDING_COUNTS,
        payload: {
          acks: wsClient.current.getPendingAcksCount(),
          responses: wsClient.current.getPendingResponsesCount()
        }
      });
    }
  }, []);

  // Funciones de envío
  const sendChatMessage = useCallback(async (text, targetUserId) => {
    if (!wsClient.current) throw new Error('WebSocket no inicializado');
    
    const tempMessage = {
      id: `temp_${Date.now()}`,
      text,
      targetUserId,
      timestamp: new Date().toISOString(),
      type: 'sending'
    };
    
    dispatch({ type: actionTypes.ADD_CHAT_MESSAGE, payload: tempMessage });
    
    try {
      const response = await wsClient.current.sendChatMessage(text, targetUserId);
      
      // Actualizar mensaje como enviado
      dispatch({
        type: actionTypes.ADD_CHAT_MESSAGE,
        payload: {
          ...tempMessage,
          id: response.pid,
          type: 'sent'
        }
      });
      
      updatePendingCounts();
      return response;
    } catch (error) {
      // Actualizar mensaje como fallido
      dispatch({
        type: actionTypes.ADD_CHAT_MESSAGE,
        payload: {
          ...tempMessage,
          type: 'failed',
          error: error.message
        }
      });
      throw error;
    }
  }, [updatePendingCounts]);

  const sendPresenceUpdate = useCallback((status, targetUserId) => {
    if (!wsClient.current) throw new Error('WebSocket no inicializado');
    wsClient.current.sendPresenceUpdate(status, targetUserId);
  }, []);

  const searchJobs = useCallback(async (query) => {
    if (!wsClient.current) throw new Error('WebSocket no inicializado');
    
    try {
      const response = await wsClient.current.sendJobSearchRequest(query);
      updatePendingCounts();
      return response;
    } catch (error) {
      dispatch({
        type: actionTypes.ADD_NOTIFICATION,
        payload: {
          id: Date.now(),
          type: 'error',
          message: `Error en búsqueda: ${error.message}`,
          timestamp: new Date().toISOString()
        }
      });
      throw error;
    }
  }, [updatePendingCounts]);

  const sendGenericRequest = useCallback(async (payload) => {
    if (!wsClient.current) throw new Error('WebSocket no inicializado');
    
    try {
      const response = await wsClient.current.sendGenericRequest(payload);
      updatePendingCounts();
      return response;
    } catch (error) {
      updatePendingCounts();
      throw error;
    }
  }, [updatePendingCounts]);

  const disconnect = useCallback(() => {
    if (wsClient.current) {
      wsClient.current.disconnect();
      isInitialized.current = false;
    }
  }, []);

  const clearNotifications = useCallback(() => {
    dispatch({ type: actionTypes.CLEAR_NOTIFICATIONS });
  }, []);

  // Cleanup al desmontar
  useEffect(() => {
    return () => {
      if (wsClient.current) {
        wsClient.current.destroy();
      }
    };
  }, []);

  const value = {
    // Estado
    ...state,
    
    // Funciones
    initializeWebSocket,
    sendChatMessage,
    sendPresenceUpdate,
    searchJobs,
    sendGenericRequest,
    disconnect,
    clearNotifications,
    updatePendingCounts
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
}

// Hook para usar el context
export function useWebSocketContext() {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocketContext debe usarse dentro de WebSocketProvider');
  }
  return context;
}
```

## Hooks Personalizados

Crea hooks específicos para diferentes funcionalidades:

### Hook para Chat
```javascript
// hooks/useChat.js
import { useCallback } from 'react';
import { useWebSocketContext } from '../contexts/WebSocketContext';

export function useChat() {
  const {
    chatMessages,
    sendChatMessage,
    isConnected,
    sendPresenceUpdate
  } = useWebSocketContext();

  const sendMessage = useCallback(async (text, targetUserId) => {
    if (!text.trim()) return;
    return await sendChatMessage(text.trim(), targetUserId);
  }, [sendChatMessage]);

  const startTyping = useCallback((targetUserId) => {
    sendPresenceUpdate('typing', targetUserId);
  }, [sendPresenceUpdate]);

  const stopTyping = useCallback((targetUserId) => {
    sendPresenceUpdate('idle', targetUserId);
  }, [sendPresenceUpdate]);

  return {
    messages: chatMessages,
    sendMessage,
    startTyping,
    stopTyping,
    isConnected
  };
}
```

### Hook para Búsqueda de Empleos
```javascript
// hooks/useJobSearch.js
import { useState, useCallback } from 'react';
import { useWebSocketContext } from '../contexts/WebSocketContext';

export function useJobSearch() {
  const { jobSearchResults, searchJobs, isConnected } = useWebSocketContext();
  const [isLoading, setIsLoading] = useState(false);

  const search = useCallback(async (query) => {
    if (!query.trim()) return;
    
    setIsLoading(true);
    try {
      await searchJobs({ query: query.trim() });
    } catch (error) {
      console.error('Error en búsqueda:', error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  }, [searchJobs]);

  return {
    results: jobSearchResults,
    search,
    isLoading,
    isConnected
  };
}
```

### Hook para Presencia
```javascript
// hooks/usePresence.js
import { useWebSocketContext } from '../contexts/WebSocketContext';

export function usePresence() {
  const { presenceUsers, sendPresenceUpdate, isConnected } = useWebSocketContext();

  return {
    presenceUsers,
    updatePresence: sendPresenceUpdate,
    isConnected
  };
}
```

### Hook para Notificaciones
```javascript
// hooks/useNotifications.js
import { useWebSocketContext } from '../contexts/WebSocketContext';

export function useNotifications() {
  const { notifications, clearNotifications } = useWebSocketContext();

  return {
    notifications,
    clearNotifications,
    hasNotifications: notifications.length > 0
  };
}
```

## Implementación en Pantallas

### Pantalla de Chat
```javascript
// screens/ChatScreen.js
import React, { useState, useEffect, useRef } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  FlatList,
  StyleSheet,
  Alert
} from 'react-native';
import { useChat } from '../hooks/useChat';

export default function ChatScreen({ route }) {
  const { targetUserId } = route.params;
  const { messages, sendMessage, startTyping, stopTyping, isConnected } = useChat();
  const [inputText, setInputText] = useState('');
  const [isSending, setIsSending] = useState(false);
  const typingTimer = useRef(null);

  // Filtrar mensajes para este chat
  const chatMessages = messages.filter(
    msg => msg.targetUserId === targetUserId || msg.fromUserId === targetUserId
  );

  const handleSendMessage = async () => {
    if (!inputText.trim() || isSending) return;

    setIsSending(true);
    try {
      await sendMessage(inputText, targetUserId);
      setInputText('');
    } catch (error) {
      Alert.alert('Error', 'No se pudo enviar el mensaje');
    } finally {
      setIsSending(false);
    }
  };

  const handleTextChange = (text) => {
    setInputText(text);
    
    // Manejar typing indicator
    startTyping(targetUserId);
    
    if (typingTimer.current) {
      clearTimeout(typingTimer.current);
    }
    
    typingTimer.current = setTimeout(() => {
      stopTyping(targetUserId);
    }, 2000);
  };

  useEffect(() => {
    return () => {
      if (typingTimer.current) {
        clearTimeout(typingTimer.current);
        stopTyping(targetUserId);
      }
    };
  }, [targetUserId, stopTyping]);

  const renderMessage = ({ item }) => (
    <View style={[
      styles.messageContainer,
      item.type === 'sent' ? styles.sentMessage : styles.receivedMessage
    ]}>
      <Text style={styles.messageText}>{item.text}</Text>
      <Text style={styles.timestamp}>
        {new Date(item.timestamp).toLocaleTimeString()}
      </Text>
      {item.type === 'failed' && (
        <Text style={styles.errorText}>Error: {item.error}</Text>
      )}
    </View>
  );

  return (
    <View style={styles.container}>
      {!isConnected && (
        <View style={styles.disconnectedBanner}>
          <Text style={styles.disconnectedText}>Desconectado</Text>
        </View>
      )}
      
      <FlatList
        data={chatMessages}
        renderItem={renderMessage}
        keyExtractor={item => item.id}
        style={styles.messagesList}
      />
      
      <View style={styles.inputContainer}>
        <TextInput
          style={styles.textInput}
          value={inputText}
          onChangeText={handleTextChange}
          placeholder="Escribe un mensaje..."
          multiline
          maxLength={500}
        />
        <TouchableOpacity
          style={[styles.sendButton, (!inputText.trim() || isSending) && styles.sendButtonDisabled]}
          onPress={handleSendMessage}
          disabled={!inputText.trim() || isSending}
        >
          <Text style={styles.sendButtonText}>
            {isSending ? 'Enviando...' : 'Enviar'}
          </Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff'
  },
  disconnectedBanner: {
    backgroundColor: '#ff6b6b',
    padding: 8,
    alignItems: 'center'
  },
  disconnectedText: {
    color: 'white',
    fontWeight: 'bold'
  },
  messagesList: {
    flex: 1,
    padding: 16
  },
  messageContainer: {
    marginVertical: 4,
    padding: 12,
    borderRadius: 8,
    maxWidth: '80%'
  },
  sentMessage: {
    backgroundColor: '#007AFF',
    alignSelf: 'flex-end'
  },
  receivedMessage: {
    backgroundColor: '#E5E5EA',
    alignSelf: 'flex-start'
  },
  messageText: {
    fontSize: 16,
    color: '#000'
  },
  timestamp: {
    fontSize: 12,
    color: '#666',
    marginTop: 4
  },
  errorText: {
    fontSize: 12,
    color: '#ff6b6b',
    marginTop: 4
  },
  inputContainer: {
    flexDirection: 'row',
    padding: 16,
    borderTopWidth: 1,
    borderTopColor: '#E5E5EA'
  },
  textInput: {
    flex: 1,
    borderWidth: 1,
    borderColor: '#E5E5EA',
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingVertical: 8,
    marginRight: 8,
    maxHeight: 100
  },
  sendButton: {
    backgroundColor: '#007AFF',
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingVertical: 8,
    justifyContent: 'center'
  },
  sendButtonDisabled: {
    backgroundColor: '#ccc'
  },
  sendButtonText: {
    color: 'white',
    fontWeight: 'bold'
  }
});
```

### Pantalla de Búsqueda de Empleos
```javascript
// screens/JobSearchScreen.js
import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  FlatList,
  ActivityIndicator,
  StyleSheet,
  Alert
} from 'react-native';
import { useJobSearch } from '../hooks/useJobSearch';

export default function JobSearchScreen() {
  const { results, search, isLoading, isConnected } = useJobSearch();
  const [query, setQuery] = useState('');

  const handleSearch = async () => {
    if (!query.trim()) {
      Alert.alert('Error', 'Ingresa un término de búsqueda');
      return;
    }

    try {
      await search(query);
    } catch (error) {
      Alert.alert('Error', 'No se pudo realizar la búsqueda');
    }
  };

  const renderJob = ({ item }) => (
    <View style={styles.jobCard}>
      <Text style={styles.jobTitle}>{item.title}</Text>
      <Text style={styles.jobCompany}>{item.company}</Text>
      <Text style={styles.jobDescription}>{item.description}</Text>
      <Text style={styles.jobSalary}>Salario: {item.salary}</Text>
    </View>
  );

  return (
    <View style={styles.container}>
      {!isConnected && (
        <View style={styles.disconnectedBanner}>
          <Text style={styles.disconnectedText}>Sin conexión</Text>
        </View>
      )}

      <View style={styles.searchContainer}>
        <TextInput
          style={styles.searchInput}
          value={query}
          onChangeText={setQuery}
          placeholder="Buscar empleos..."
          onSubmitEditing={handleSearch}
        />
        <TouchableOpacity
          style={[styles.searchButton, (!isConnected || isLoading) && styles.searchButtonDisabled]}
          onPress={handleSearch}
          disabled={!isConnected || isLoading}
        >
          {isLoading ? (
            <ActivityIndicator color="white" />
          ) : (
            <Text style={styles.searchButtonText}>Buscar</Text>
          )}
        </TouchableOpacity>
      </View>

      <FlatList
        data={results}
        renderItem={renderJob}
        keyExtractor={item => item.id}
        style={styles.resultsList}
        ListEmptyComponent={
          <Text style={styles.emptyText}>
            {results.length === 0 && !isLoading ? 'No hay resultados' : ''}
          </Text>
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff'
  },
  disconnectedBanner: {
    backgroundColor: '#ff6b6b',
    padding: 8,
    alignItems: 'center'
  },
  disconnectedText: {
    color: 'white',
    fontWeight: 'bold'
  },
  searchContainer: {
    flexDirection: 'row',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#E5E5EA'
  },
  searchInput: {
    flex: 1,
    borderWidth: 1,
    borderColor: '#E5E5EA',
    borderRadius: 8,
    paddingHorizontal: 12,
    paddingVertical: 8,
    marginRight: 8
  },
  searchButton: {
    backgroundColor: '#007AFF',
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 8,
    justifyContent: 'center'
  },
  searchButtonDisabled: {
    backgroundColor: '#ccc'
  },
  searchButtonText: {
    color: 'white',
    fontWeight: 'bold'
  },
  resultsList: {
    flex: 1,
    padding: 16
  },
  jobCard: {
    backgroundColor: '#f9f9f9',
    borderRadius: 8,
    padding: 16,
    marginBottom: 8,
    borderWidth: 1,
    borderColor: '#E5E5EA'
  },
  jobTitle: {
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 4
  },
  jobCompany: {
    fontSize: 16,
    color: '#666',
    marginBottom: 8
  },
  jobDescription: {
    fontSize: 14,
    marginBottom: 8
  },
  jobSalary: {
    fontSize: 14,
    fontWeight: 'bold',
    color: '#007AFF'
  },
  emptyText: {
    textAlign: 'center',
    color: '#666',
    marginTop: 32
  }
});
```

## Ejemplos Prácticos

### Inicialización en App.js
```javascript
// App.js
import React, { useEffect } from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import AsyncStorage from '@react-native-async-storage/async-storage';

import { WebSocketProvider, useWebSocketContext } from './contexts/WebSocketContext';
import LoginScreen from './screens/LoginScreen';
import ChatScreen from './screens/ChatScreen';
import JobSearchScreen from './screens/JobSearchScreen';

const Stack = createStackNavigator();

function AppNavigator() {
  const { initializeWebSocket, isConnected } = useWebSocketContext();

  useEffect(() => {
    initializeApp();
  }, []);

  const initializeApp = async () => {
    try {
      const token = await AsyncStorage.getItem('authToken');
      if (token) {
        await initializeWebSocket({
          url: 'ws://tu-servidor.com/ws',
          token: token
        });
      }
    } catch (error) {
      console.error('Error inicializando app:', error);
    }
  };

  return (
    <NavigationContainer>
      <Stack.Navigator>
        <Stack.Screen name="Login" component={LoginScreen} />
        <Stack.Screen name="Chat" component={ChatScreen} />
        <Stack.Screen name="JobSearch" component={JobSearchScreen} />
      </Stack.Navigator>
    </NavigationContainer>
  );
}

export default function App() {
  return (
    <WebSocketProvider>
      <AppNavigator />
    </WebSocketProvider>
  );
}
```

### Hook de Conexión Automática
```javascript
// hooks/useAutoConnect.js
import { useEffect, useRef } from 'react';
import { AppState } from 'react-native';
import { useWebSocketContext } from '../contexts/WebSocketContext';
import AsyncStorage from '@react-native-async-storage/async-storage';

export function useAutoConnect() {
  const { initializeWebSocket, disconnect, connectionState } = useWebSocketContext();
  const appState = useRef(AppState.currentState);

  useEffect(() => {
    const handleAppStateChange = async (nextAppState) => {
      if (appState.current.match(/inactive|background/) && nextAppState === 'active') {
        // App volvió al primer plano, reconectar si es necesario
        if (connectionState === 'disconnected' || connectionState === 'failed') {
          const token = await AsyncStorage.getItem('authToken');
          if (token) {
            initializeWebSocket({
              url: 'ws://tu-servidor.com/ws',
              token: token
            });
          }
        }
      } else if (nextAppState.match(/inactive|background/)) {
        // App va al fondo, desconectar para ahorrar batería
        disconnect();
      }
      
      appState.current = nextAppState;
    };

    const subscription = AppState.addEventListener('change', handleAppStateChange);
    return () => subscription?.remove();
  }, [initializeWebSocket, disconnect, connectionState]);
}
```

### Componente de Estado de Conexión
```javascript
// components/ConnectionStatus.js
import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { useWebSocketContext } from '../contexts/WebSocketContext';

export default function ConnectionStatus() {
  const { connectionState, pendingAcks, pendingResponses } = useWebSocketContext();

  const getStatusColor = () => {
    switch (connectionState) {
      case 'connected': return '#4CAF50';
      case 'connecting': return '#FF9800';
      case 'reconnecting': return '#FF9800';
      case 'disconnected': return '#757575';
      case 'failed': return '#F44336';
      default: return '#757575';
    }
  };

  const getStatusText = () => {
    switch (connectionState) {
      case 'connected': return 'Conectado';
      case 'connecting': return 'Conectando...';
      case 'reconnecting': return 'Reconectando...';
      case 'disconnected': return 'Desconectado';
      case 'failed': return 'Error de conexión';
      default: return 'Desconocido';
    }
  };

  return (
    <View style={styles.container}>
      <View style={[styles.indicator, { backgroundColor: getStatusColor() }]} />
      <Text style={styles.statusText}>{getStatusText()}</Text>
      {(pendingAcks > 0 || pendingResponses > 0) && (
        <Text style={styles.pendingText}>
          Pendientes: {pendingAcks + pendingResponses}
        </Text>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 8,
    backgroundColor: '#f0f0f0'
  },
  indicator: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginRight: 8
  },
  statusText: {
    fontSize: 12,
    color: '#333'
  },
  pendingText: {
    fontSize: 10,
    color: '#666',
    marginLeft: 8
  }
});
```

## Mejores Prácticas

### 1. Gestión de Memoria
```javascript
// Limpiar listeners y timers al desmontar componentes
useEffect(() => {
  return () => {
    // Cleanup
  };
}, []);
```

### 2. Manejo de Errores
```javascript
const handleWebSocketError = (error) => {
  console.error('WebSocket error:', error);
  
  // Mostrar error al usuario si es crítico
  if (error.type === 'auth_failed') {
    Alert.alert('Error', 'Tu sesión ha expirado. Inicia sesión nuevamente.');
    // Redirigir a login
  }
};
```

### 3. Optimización de Re-renders
```javascript
// Usar React.memo para componentes que reciben datos del WebSocket
const ChatMessage = React.memo(({ message }) => {
  return (
    <View>
      <Text>{message.text}</Text>
    </View>
  );
});
```

### 4. Persistencia de Datos
```javascript
// Guardar mensajes importantes en AsyncStorage
const saveMessageToStorage = async (message) => {
  try {
    const savedMessages = await AsyncStorage.getItem('chatMessages');
    const messages = savedMessages ? JSON.parse(savedMessages) : [];
    messages.push(message);
    await AsyncStorage.setItem('chatMessages', JSON.stringify(messages));
  } catch (error) {
    console.error('Error saving message:', error);
  }
};
```

### 5. Throttling y Debouncing
```javascript
// Para typing indicators
const debouncedStopTyping = useCallback(
  debounce((targetUserId) => {
    stopTyping(targetUserId);
  }, 2000),
  [stopTyping]
);
```

## Troubleshooting

### Problemas Comunes

#### 1. WebSocket no se conecta
```javascript
// Verificar configuración
console.log('WebSocket URL:', config.url);
console.log('Auth token:', config.token);

// Verificar conectividad de red
import NetInfo from '@react-native-netinfo/netinfo';

NetInfo.fetch().then(state => {
  console.log('Connection type', state.type);
  console.log('Is connected?', state.isConnected);
});
```

#### 2. Reconexión infinita
```javascript
// Limitar intentos de reconexión
const maxReconnectAttempts = 5;
const reconnectInterval = 5000; // Incrementar tiempo entre intentos
```

#### 3. Mensajes duplicados
```javascript
// Usar IDs únicos y verificar duplicados
const isDuplicate = (newMessage, existingMessages) => {
  return existingMessages.some(msg => msg.id === newMessage.id);
};
```

#### 4. Memory leaks
```javascript
// Limpiar subscripciones y timers
useEffect(() => {
  return () => {
    if (timer) clearTimeout(timer);
    if (subscription) subscription.remove();
  };
}, []);
```

### Debugging

#### Log detallado
```javascript
const debugLog = (component, action, data) => {
  if (__DEV__) {
    console.log(`[${component}] ${action}:`, data);
  }
};
```

#### Herramientas de desarrollo
```javascript
// Flipper plugin para WebSocket debugging
import { logger } from 'flipper';

logger.info('WebSocket message', { type: message.type, pid: message.pid });
```

Esta guía proporciona una implementación completa y eficiente para usar WebSockets en React Native con Expo, optimizada para múltiples pantallas con una sintaxis simple y mantenible. 