import type { NetworkInterface, PacketFilterRule } from "./types";

const base =  import.meta.env.VITE_API_BASE ?? "http://localhost:8080";

export async function fetchInterfaces(): Promise<NetworkInterface[]> {
  console.log(`${base}/interfaces`);

  const res = await fetch(`${base}/interfaces`);

  if (!res.ok) {
    throw new Error(await res.text());
  }

  const data = await res.json();

  console.log(data);

  return data as NetworkInterface[];
}
export async function startCapture(interfaceId: number): Promise<void> {
  const res = await fetch(`${base}/capture/start/${interfaceId}`, {
    method: "POST",
  });
  if (!res.ok) throw new Error(await res.text());
}

export async function stopCapture(): Promise<void> {
  const res = await fetch(`${base}/capture/stop`, { method: "POST" });
  if (!res.ok) throw new Error(await res.text());
}

export async function updateFilter(rule: PacketFilterRule): Promise<void> {
  const res = await fetch(`${base}/filter`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(rule),
  });
  if (!res.ok) throw new Error(await res.text());
}
