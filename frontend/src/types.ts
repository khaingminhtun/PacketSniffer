export type NetworkInterface = {
  id: number;
  pcap_id: string;
  name: string;
  description: string;
  type: string;
  ips: string[];
  is_loopback: boolean;
  is_virtual: boolean;
  is_up: boolean;
};

export type PacketPayload = {
  Timestamp: string;
  SrcIP: string;
  DstIP: string;
  SrcPort: number;
  DstPort: number;
  Protocol: string;
  AppProtocol: string;
  TCPFlags: string;
  IsDNS: boolean;
  DNSQuery: string;
  Length: number;
  TLSSNI: string;
  IsTLS: boolean;
};

export type StreamPayload = {
  Key: string;
  SrcIP: string;
  DstIP: string;
  SrcPort: string;
  DstPort: string;
  Bytes?: number[] | null;
  DataLen: number;
  IsTLS: boolean;
  SNI: string;
};

export type BusPacketEvent = { type: "packet"; data: PacketPayload };
export type BusStreamEvent = { type: "stream"; data: StreamPayload };
export type HandshakeMessage = { message: string };

export type WsPayload = BusPacketEvent | BusStreamEvent | HandshakeMessage;

export function isBusPacketEvent(m: WsPayload): m is BusPacketEvent {
  return "type" in m && m.type === "packet" && "data" in m;
}

export function isBusStreamEvent(m: WsPayload): m is BusStreamEvent {
  return "type" in m && m.type === "stream" && "data" in m;
}

export function isHandshake(m: WsPayload): m is HandshakeMessage {
  return "message" in m && !("type" in m);
}

export type PacketFilterRule = {
  allowed_protocols: string[];
  blocked_ports: number[];
};
