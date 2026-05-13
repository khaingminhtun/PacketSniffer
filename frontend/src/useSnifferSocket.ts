import { useCallback, useEffect, useRef, useState } from "react";
import type { BusPacketEvent, BusStreamEvent, WsPayload } from "./types";
import {
  isBusPacketEvent,
  isBusStreamEvent,
  isHandshake,
} from "./types";

export type ConnectionState = "connecting" | "open" | "closed" | "error";

const MAX_PACKETS = 400;
const MAX_STREAMS = 200;

function wsUrl(): string {
  const fromEnv = import.meta.env.VITE_WS_URL as string | undefined;
  if (fromEnv) return fromEnv;
  const proto = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${proto}//${window.location.host}/ws`;
}

export function useSnifferSocket() {
  const [state, setState] = useState<ConnectionState>("connecting");
  const [lastError, setLastError] = useState<string | null>(null);
  const [handshake, setHandshake] = useState<string | null>(null);
  const [packets, setPackets] = useState<BusPacketEvent["data"][]>([]);
  const [streams, setStreams] = useState<BusStreamEvent["data"][]>([]);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const shouldReconnect = useRef(true);

  const pushPacket = useCallback((p: BusPacketEvent["data"]) => {
    setPackets((prev) => {
      const next = [p, ...prev];
      return next.slice(0, MAX_PACKETS);
    });
  }, []);

  const pushStream = useCallback((s: BusStreamEvent["data"]) => {
    setStreams((prev) => {
      const next = [s, ...prev];
      return next.slice(0, MAX_STREAMS);
    });
  }, []);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    setState("connecting");
    setLastError(null);

    const ws = new WebSocket(wsUrl());
    wsRef.current = ws;

    ws.onopen = () => {
      setState("open");
    };

    ws.onclose = () => {
      setState("closed");
      wsRef.current = null;
      if (shouldReconnect.current) {
        reconnectTimer.current = setTimeout(connect, 2000);
      }
    };

    ws.onerror = () => {
      setState("error");
      setLastError("WebSocket error");
    };

    ws.onmessage = (ev) => {
      try {
        const parsed = JSON.parse(ev.data as string) as WsPayload;
        if (isHandshake(parsed)) {
          setHandshake(parsed.message);
          return;
        }
        if (isBusPacketEvent(parsed)) {
          pushPacket(parsed.data);
          return;
        }
        if (isBusStreamEvent(parsed)) {
          pushStream(parsed.data);
        }
      } catch {
        setLastError("Invalid JSON from server");
      }
    };
  }, [pushPacket, pushStream]);

  useEffect(() => {
    shouldReconnect.current = true;
    connect();
    return () => {
      shouldReconnect.current = false;
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [connect]);

  const clearFeed = useCallback(() => {
    setPackets([]);
    setStreams([]);
  }, []);

  return {
    connection: state,
    lastError,
    handshake,
    packets,
    streams,
    clearFeed,
    reconnect: connect,
  };
}
