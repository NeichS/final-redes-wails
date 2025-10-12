export interface EventMessage {
  id: number;
  text: string;
  type: 'success' | 'error' | 'info';
}