export type ColorMode = 'light' | 'dark' | 'system';

export type AccentColor = 'indigo' | 'red' | 'blue' | 'green' | 'purple' | 'orange';

export type AppStatus =
  | 'idle'
  | 'loading'
  | 'success'
  | 'error'
  | 'empty'
  | 'server_down';

export interface User {
  id: string;
  fullName: string;
  username: string;
  email: string;
  avatarUrl?: string;
}
