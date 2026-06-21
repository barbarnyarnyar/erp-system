import React, { createContext, useContext, useState } from 'react';

interface AuthContextType {
  token: string | null;
  activeSelectedLegalEntityId: string;
  login: (token: string) => void;
  logout: () => void;
  setLegalEntityId: (id: string) => void;
  isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [token, setToken] = useState<string | null>(localStorage.getItem('erp_jwt_token'));
  const [activeSelectedLegalEntityId, setLegalEntityIdState] = useState<string>(
    localStorage.getItem('erp_active_legal_entity_id') || 'default_entity_id'
  );

  const login = (newToken: string) => {
    localStorage.setItem('erp_jwt_token', newToken);
    setToken(newToken);
  };

  const logout = () => {
    localStorage.removeItem('erp_jwt_token');
    setToken(null);
  };

  const setLegalEntityId = (id: string) => {
    localStorage.setItem('erp_active_legal_entity_id', id);
    setLegalEntityIdState(id);
  };

  const isAuthenticated = !!token;

  return (
    <AuthContext.Provider
      value={{
        token,
        activeSelectedLegalEntityId,
        login,
        logout,
        setLegalEntityId,
        isAuthenticated,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
