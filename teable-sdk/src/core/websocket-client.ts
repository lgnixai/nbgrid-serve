/**
 * WebSocket 客户端实现
 * 提供实时协作、通知推送等功能
 */

import WebSocket from 'ws';

// 简单的 EventEmitter 实现，兼容浏览器环境
class EventEmitter {
  private events: { [key: string]: Function[] } = {};

  on(event: string, listener: Function) {
    if (!this.events[event]) {
      this.events[event] = [];
    }
    this.events[event]!.push(listener);
  }

  emit(event: string, ...args: any[]) {
    if (this.events[event]) {
      this.events[event]!.forEach(listener => listener(...args));
    }
  }

  off(event: string, listener: Function) {
    if (this.events[event]) {
      this.events[event] = this.events[event]!.filter(l => l !== listener);
    }
  }
}
import { 
  WebSocketMessage, 
  CollaborationMessage, 
  RecordChangeMessage,
  TeableConfig 
} from '../types';

export interface WebSocketClientOptions {
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
  debug?: boolean;
}

export class WebSocketClient extends EventEmitter {
  private ws: WebSocket | undefined;
  private config: TeableConfig;
  private options: WebSocketClientOptions;
  private reconnectAttempts: number = 0;
  private reconnectTimer: any | undefined;
  private heartbeatTimer: any | undefined;
  private isConnecting: boolean = false;
  private isConnected: boolean = false;

  constructor(config: TeableConfig, options: WebSocketClientOptions = {}) {
    super();
    this.config = config;
    this.options = {
      reconnectInterval: 5000,
      maxReconnectAttempts: 10,
      heartbeatInterval: 30000,
      debug: false,
      ...options
    };
  }

  public async connect(): Promise<void> {
    if (this.isConnecting || this.isConnected) {
      return;
    }

    this.isConnecting = true;

    try {
      const wsUrl = this.config.baseUrl.replace(/^https?:\/\//, 'ws://').replace(/^https:\/\//, 'wss://');
      const url = `${wsUrl}/api/ws/socket?token=${this.config.accessToken}`;

      if (this.options.debug) {
        console.log('[Teable WebSocket] Connecting to:', url);
      }

      this.ws = new WebSocket(url);

      this.ws.on('open', this.handleOpen.bind(this));
      this.ws.on('message', this.handleMessage.bind(this));
      this.ws.on('close', this.handleClose.bind(this));
      this.ws.on('error', this.handleError.bind(this));

    } catch (error) {
      this.isConnecting = false;
      this.emit('error', error);
      throw error;
    }
  }

  public disconnect(): void {
    this.clearTimers();
    
    if (this.ws) {
      this.ws.close();
      this.ws = undefined;
    }
    
    this.isConnected = false;
    this.isConnecting = false;
    this.reconnectAttempts = 0;
  }

  public send(message: Partial<WebSocketMessage>): void {
    if (!this.isConnected || !this.ws) {
      throw new Error('WebSocket is not connected');
    }

    const fullMessage: WebSocketMessage = {
      ...message,
      timestamp: message.timestamp || new Date().toISOString()
    } as WebSocketMessage;

    const messageStr = JSON.stringify(fullMessage);

    if (this.options.debug) {
      console.log('[Teable WebSocket] Sending:', messageStr);
    }

    this.ws.send(messageStr);
  }

  public joinChannel(channel: string): void {
    this.send({
      type: 'join_channel',
      data: { channel },
      timestamp: new Date().toISOString()
    });
  }

  public leaveChannel(channel: string): void {
    this.send({
      type: 'leave_channel',
      data: { channel },
      timestamp: new Date().toISOString()
    });
  }

  public updatePresence(resourceType: string, resourceId: string, cursorPosition?: { x: number; y: number }): void {
    this.send({
      type: 'presence_update',
      data: {
        resource_type: resourceType,
        resource_id: resourceId,
        cursor_position: cursorPosition
      },
      timestamp: new Date().toISOString()
    });
  }

  public updateCursor(resourceType: string, resourceId: string, cursorPosition: { x: number; y: number }, fieldId?: string, recordId?: string): void {
    this.send({
      type: 'cursor_update',
      data: {
        resource_type: resourceType,
        resource_id: resourceId,
        cursor_position: cursorPosition,
        field_id: fieldId,
        record_id: recordId
      },
      timestamp: new Date().toISOString()
    });
  }

  public subscribeToTable(tableId: string): void {
    this.joinChannel(`table:${tableId}`);
  }

  public unsubscribeFromTable(tableId: string): void {
    this.leaveChannel(`table:${tableId}`);
  }

  public subscribeToRecord(tableId: string, recordId: string): void {
    this.joinChannel(`record:${tableId}:${recordId}`);
  }

  public unsubscribeFromRecord(tableId: string, recordId: string): void {
    this.leaveChannel(`record:${tableId}:${recordId}`);
  }

  public subscribeToView(viewId: string): void {
    this.joinChannel(`view:${viewId}`);
  }

  public unsubscribeFromView(viewId: string): void {
    this.leaveChannel(`view:${viewId}`);
  }

  private handleOpen(): void {
    if (this.options.debug) {
      console.log('[Teable WebSocket] Connected');
    }

    this.isConnected = true;
    this.isConnecting = false;
    this.reconnectAttempts = 0;
    
    this.startHeartbeat();
    this.emit('connected');
  }

  private handleMessage(data: WebSocket.Data): void {
    try {
      const message: WebSocketMessage = JSON.parse(data.toString());
      
      if (this.options.debug) {
        console.log('[Teable WebSocket] Received:', message);
      }

      this.emit('message', message);
      
      // 根据消息类型触发特定事件
      switch (message.type) {
        case 'collaboration':
          this.emit('collaboration', message as CollaborationMessage);
          break;
        case 'record_change':
          this.emit('record_change', message as RecordChangeMessage);
          break;
        case 'presence_update':
          this.emit('presence_update', message);
          break;
        case 'cursor_update':
          this.emit('cursor_update', message);
          break;
        case 'notification':
          this.emit('notification', message);
          break;
        case 'pong':
          // 心跳响应
          break;
        default:
          this.emit('unknown_message', message);
      }
    } catch (error) {
      if (this.options.debug) {
        console.error('[Teable WebSocket] Failed to parse message:', error);
      }
      this.emit('error', error);
    }
  }

  private handleClose(code: number, reason: string): void {
    if (this.options.debug) {
      console.log('[Teable WebSocket] Disconnected:', code, reason);
    }

    this.isConnected = false;
    this.isConnecting = false;
    this.clearTimers();
    
    this.emit('disconnected', { code, reason });

    // 如果不是主动断开，尝试重连
    if (code !== 1000 && this.reconnectAttempts < this.options.maxReconnectAttempts!) {
      this.scheduleReconnect();
    }
  }

  private handleError(error: Error): void {
    if (this.options.debug) {
      console.error('[Teable WebSocket] Error:', error);
    }
    
    this.emit('error', error);
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      return;
    }

    this.reconnectAttempts++;
    const delay = this.options.reconnectInterval! * Math.pow(2, this.reconnectAttempts - 1);

    if (this.options.debug) {
      console.log(`[Teable WebSocket] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);
    }

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = undefined;
      this.connect().catch(error => {
        if (this.options.debug) {
          console.error('[Teable WebSocket] Reconnect failed:', error);
        }
        this.emit('reconnect_failed', error);
      });
    }, delay);
  }

  private startHeartbeat(): void {
    this.heartbeatTimer = setInterval(() => {
      if (this.isConnected && this.ws) {
        this.send({ type: 'ping', data: {}, timestamp: new Date().toISOString() });
      }
    }, this.options.heartbeatInterval);
  }

  private clearTimers(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = undefined;
    }
    
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = undefined;
    }
  }

  public getConnectionState(): 'connecting' | 'connected' | 'disconnected' {
    if (this.isConnecting) return 'connecting';
    if (this.isConnected) return 'connected';
    return 'disconnected';
  }

  public getReconnectAttempts(): number {
    return this.reconnectAttempts;
  }
}
