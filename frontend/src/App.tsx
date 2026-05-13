import { type FormEvent, useEffect, useMemo, useState } from "react";
import {
  fetchInterfaces,
  startCapture,
  stopCapture,
  updateFilter,
} from "./api";
import type { NetworkInterface, PacketFilterRule } from "./types";
import { useSnifferSocket } from "./useSnifferSocket";

function statusDot(connection: string) {
  switch (connection) {
    case "open":
      return { label: "Live", color: "var(--good)" };
    case "connecting":
      return { label: "Connecting…", color: "var(--warn)" };
    case "error":
      return { label: "Error", color: "var(--bad)" };
    default:
      return { label: "Disconnected", color: "var(--muted)" };
  }
}

export default function App() {
  const {
    connection,
    lastError,
    handshake,
    packets,
    streams,
    clearFeed,
  } = useSnifferSocket();

  const [ifaces, setIfaces] = useState<NetworkInterface[]>([]);
  const [selectedId, setSelectedId] = useState<number | "">("");
  const [loadingIfaces, setLoadingIfaces] = useState(true);
  const [actionError, setActionError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const [protocolsInput, setProtocolsInput] = useState("TCP,UDP");
  const [blockedPortsInput, setBlockedPortsInput] = useState("");

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const list = await fetchInterfaces();
        if (!cancelled) {
          setIfaces(list);
          const firstReal = list.find((i) => !i.is_loopback && i.is_up);
          setSelectedId(firstReal?.id ?? list[0]?.id ?? "");
        }
      } catch (e) {
        if (!cancelled) {
          setActionError(e instanceof Error ? e.message : "Failed to load interfaces");
        }
      } finally {
        if (!cancelled) setLoadingIfaces(false);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const dot = useMemo(() => statusDot(connection), [connection]);

  async function onStart() {
    if (selectedId === "") return;
    setBusy(true);
    setActionError(null);
    try {
      await startCapture(selectedId);
    } catch (e) {
      setActionError(e instanceof Error ? e.message : "Start failed");
    } finally {
      setBusy(false);
    }
  }

  async function onStop() {
    setBusy(true);
    setActionError(null);
    try {
      await stopCapture();
    } catch (e) {
      setActionError(e instanceof Error ? e.message : "Stop failed");
    } finally {
      setBusy(false);
    }
  }

  async function onApplyFilter(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setActionError(null);
    const allowed_protocols = protocolsInput
      .split(/[,;\s]+/)
      .map((s) => s.trim().toUpperCase())
      .filter(Boolean);
    const blocked_ports = blockedPortsInput
      .split(/[,;\s]+/)
      .map((s) => parseInt(s.trim(), 10))
      .filter((n) => !Number.isNaN(n));
    const rule: PacketFilterRule = { allowed_protocols, blocked_ports };
    try {
      await updateFilter(rule);
    } catch (err) {
      setActionError(err instanceof Error ? err.message : "Filter update failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div style={{ display: "flex", flexDirection: "column", minHeight: "100vh" }}>
      <header
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          padding: "0.75rem 1.25rem",
          borderBottom: "1px solid var(--border)",
          background: "var(--panel)",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: "0.75rem" }}>
          <strong style={{ fontSize: "1.05rem" }}>Packet Sniffer</strong>
          <span
            className="mono"
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: "0.35rem",
              color: "var(--muted)",
              fontSize: "0.85rem",
            }}
          >
            <span
              aria-hidden
              style={{
                width: 8,
                height: 8,
                borderRadius: "50%",
                background: dot.color,
                display: "inline-block",
              }}
            />
            {dot.label}
          </span>
          {handshake ? (
            <span className="mono" style={{ color: "var(--muted)", fontSize: "0.8rem" }}>
              {handshake}
            </span>
          ) : null}
        </div>
        <button type="button" onClick={clearFeed} disabled={busy}>
          Clear feed
        </button>
      </header>

      {(actionError || lastError) && (
        <div
          role="alert"
          style={{
            padding: "0.5rem 1.25rem",
            background: "#3a2226",
            color: "#ffb4bc",
            fontSize: "0.9rem",
          }}
        >
          {actionError ?? lastError}
        </div>
      )}

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "280px 1fr 1fr",
          gap: "1rem",
          padding: "1rem",
          flex: 1,
          minHeight: 0,
        }}
      >
        <aside
          style={{
            display: "flex",
            flexDirection: "column",
            gap: "1rem",
            padding: "1rem",
            background: "var(--panel)",
            borderRadius: 10,
            border: "1px solid var(--border)",
            height: "fit-content",
          }}
        >
          <h2 style={{ margin: 0, fontSize: "0.95rem", color: "var(--muted)" }}>
            Capture
          </h2>
          <label style={{ display: "flex", flexDirection: "column", gap: "0.35rem" }}>
            <span style={{ fontSize: "0.8rem", color: "var(--muted)" }}>Interface</span>
            <select
              value={selectedId === "" ? "" : String(selectedId)}
              onChange={(e) =>
                setSelectedId(e.target.value === "" ? "" : Number(e.target.value))
              }
              disabled={loadingIfaces || ifaces.length === 0}
            >
              {ifaces.length === 0 ? (
                <option value="">No interfaces</option>
              ) : (
                ifaces.map((i) => (
                  <option key={i.id} value={i.id}>
                    {i.name} {i.is_loopback ? "(loopback)" : ""}
                  </option>
                ))
              )}
            </select>
          </label>
          <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap" }}>
            <button type="button" className="primary" onClick={onStart} disabled={busy || selectedId === ""}>
              Start
            </button>
            <button type="button" onClick={onStop} disabled={busy}>
              Stop
            </button>
          </div>

          <form onSubmit={onApplyFilter} style={{ display: "flex", flexDirection: "column", gap: "0.6rem" }}>
            <h2 style={{ margin: 0, fontSize: "0.95rem", color: "var(--muted)" }}>
              Filter
            </h2>
            <label style={{ display: "flex", flexDirection: "column", gap: "0.35rem" }}>
              <span style={{ fontSize: "0.8rem", color: "var(--muted)" }}>
                Allowed protocols
              </span>
              <input
                value={protocolsInput}
                onChange={(e) => setProtocolsInput(e.target.value)}
                placeholder="TCP, UDP"
              />
            </label>
            <label style={{ display: "flex", flexDirection: "column", gap: "0.35rem" }}>
              <span style={{ fontSize: "0.8rem", color: "var(--muted)" }}>
                Blocked ports
              </span>
              <input
                value={blockedPortsInput}
                onChange={(e) => setBlockedPortsInput(e.target.value)}
                placeholder="443, 80"
              />
            </label>
            <button type="submit" className="primary" disabled={busy}>
              Apply filter
            </button>
          </form>
        </aside>

        <section
          style={{
            display: "flex",
            flexDirection: "column",
            minHeight: 0,
            background: "var(--panel)",
            borderRadius: 10,
            border: "1px solid var(--border)",
            overflow: "hidden",
          }}
        >
          <div style={{ padding: "0.65rem 1rem", borderBottom: "1px solid var(--border)" }}>
            <strong>Packets</strong>
            <span style={{ color: "var(--muted)", marginLeft: "0.5rem", fontSize: "0.85rem" }}>
              {packets.length} rows
            </span>
          </div>
          <div style={{ overflow: "auto", flex: 1, maxHeight: "calc(100vh - 8rem)" }}>
            <table>
              <thead>
                <tr>
                  <th>Time</th>
                  <th>Src</th>
                  <th>Dst</th>
                  <th>Proto</th>
                  <th>App</th>
                  <th>Len</th>
                </tr>
              </thead>
              <tbody>
                {packets.map((p, idx) => (
                  <tr key={`${p.Timestamp}-${idx}`}>
                    <td className="mono">{p.Timestamp.slice(11, 23)}</td>
                    <td className="mono">
                      {p.SrcIP}:{p.SrcPort}
                    </td>
                    <td className="mono">
                      {p.DstIP}:{p.DstPort}
                    </td>
                    <td>{p.Protocol}</td>
                    <td>
                      {p.AppProtocol}
                      {p.IsTLS && p.TLSSNI ? (
                        <span style={{ color: "var(--muted)" }}> · {p.TLSSNI}</span>
                      ) : null}
                      {p.IsDNS && p.DNSQuery ? (
                        <span style={{ color: "var(--muted)" }}> · {p.DNSQuery}</span>
                      ) : null}
                    </td>
                    <td>{p.Length}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            {packets.length === 0 ? (
              <p style={{ padding: "1rem", color: "var(--muted)", margin: 0 }}>
                Waiting for packets. Start capture on the server interface.
              </p>
            ) : null}
          </div>
        </section>

        <section
          style={{
            display: "flex",
            flexDirection: "column",
            minHeight: 0,
            background: "var(--panel)",
            borderRadius: 10,
            border: "1px solid var(--border)",
            overflow: "hidden",
          }}
        >
          <div style={{ padding: "0.65rem 1rem", borderBottom: "1px solid var(--border)" }}>
            <strong>Streams</strong>
            <span style={{ color: "var(--muted)", marginLeft: "0.5rem", fontSize: "0.85rem" }}>
              {streams.length} rows
            </span>
          </div>
          <div style={{ overflow: "auto", flex: 1, maxHeight: "calc(100vh - 8rem)" }}>
            <table>
              <thead>
                <tr>
                  <th>Flow</th>
                  <th>Bytes</th>
                  <th>TLS / SNI</th>
                </tr>
              </thead>
              <tbody>
                {streams.map((s, idx) => (
                  <tr key={`${s.Key}-${idx}`}>
                    <td className="mono">
                      {s.SrcIP}:{s.SrcPort} → {s.DstIP}:{s.DstPort}
                    </td>
                    <td>{s.DataLen}</td>
                    <td>
                      {s.IsTLS ? "yes" : "—"}
                      {s.SNI ? (
                        <span style={{ color: "var(--muted)" }}> {s.SNI}</span>
                      ) : null}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {streams.length === 0 ? (
              <p style={{ padding: "1rem", color: "var(--muted)", margin: 0 }}>
                Reassembled stream events appear here (including TLS SNI when detected).
              </p>
            ) : null}
          </div>
        </section>
      </div>
    </div>
  );
}
