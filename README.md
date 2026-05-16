# Packet Sniffer

Real-time packet capture and inspection in the browser. The **Go** backend captures traffic on your machine with **libpcap** (gopacket); the **React** frontend shows live packets and streams over WebSocket.

Run it **on the host** (not in Docker) so the sniffer sees the same network interfaces and traffic as your PC. A Linux container only sees the container network, not packets on your Wi‑Fi or Ethernet adapter.

## Quick start

You need two terminals after cloning.

### 1. Clone

```bash
git clone <your-repo-url>
cd PacketSniffer
```

### 2. Backend (terminal 1)

```bash
cd backend
go run ./cmd/sniffer
```

API and WebSocket listen on **http://localhost:8080**.

### 3. Frontend (terminal 2)

```bash
cd frontend
npm install
npm run dev
```

Open **http://localhost:5173**, choose an interface, then **Start capture**.

---

## Requirements

| Tool | Version |
|------|---------|
| [Go](https://go.dev/dl/) | 1.24+ (see `backend/go.mod`) |
| [Node.js](https://nodejs.org/) | 20+ |
| Packet capture library | Platform-specific (below) |
| C compiler | Required for Go + libpcap (**CGO**) |

### Windows

1. Install **[Npcap](https://npcap.com/)** (WinPcap-compatible mode during setup is fine).
2. Install a C toolchain for CGO, e.g. [MSYS2](https://www.msys2.org/) with `pacman -S mingw-w64-x86_64-gcc`, and add `mingw64\bin` to your `PATH`.
3. Run the backend **as Administrator** so capture can open adapters.

```powershell
cd backend
$env:CGO_ENABLED="1"
go run ./cmd/sniffer
```

### Linux

```bash
sudo apt install -y gcc libpcap-dev   # Debian/Ubuntu
# or: sudo dnf install gcc libpcap-devel   # Fedora

cd backend
go run ./cmd/sniffer
```

Capture often needs elevated privileges (`sudo go run ...`) unless your user is in the `pcap` group.

### macOS

```bash
xcode-select --install
cd backend
go run ./cmd/sniffer
```

libpcap is included; you may be prompted for permission to capture.

---

## Architecture & design flow

The app splits into three layers: **UI** (React), **API + realtime** (HTTP + WebSocket), and **capture pipeline** (libpcap → parse → filter → bus). Control commands go in over HTTP; live data comes back over WebSocket.

### System overview

```mermaid
flowchart TB
  subgraph Frontend["Frontend (React + Vite :5173)"]
    UI[App.tsx]
    API_CLIENT[api.ts — REST]
    WS_HOOK[useSnifferSocket.ts — WebSocket]
    UI --> API_CLIENT
    UI --> WS_HOOK
  end

  subgraph Backend["Backend (Go :8080)"]
    HTTP[HTTP routes + CORS]
    HUB[WebSocket Hub]
    BC[Broadcaster]
    IF_SVC[InterfaceService]
    CAP_SVC[CaptureService]
    BUS[EventBus]

    HTTP --> IF_SVC
    HTTP --> CAP_SVC
    CAP_SVC --> BUS
    BUS --> BC --> HUB
  end

  subgraph Pipeline["Capture pipeline (goroutines)"]
    PCAP[libpcap / gopacket]
    FAN[Fan-out]
  end

  NIC[(Host network adapter)]

  API_CLIENT -->|HTTP proxy| HTTP
  WS_HOOK -->|WS proxy /ws| HUB
  CAP_SVC --> PCAP
  PCAP --> NIC
  PCAP --> FAN
```

### Control flow (what the user triggers)

```mermaid
sequenceDiagram
  actor User
  participant UI as React UI
  participant API as Go HTTP API
  participant IF as InterfaceService
  participant CAP as CaptureService
  participant WS as WebSocket Hub

  User->>UI: Open dashboard
  UI->>API: GET /interfaces
  API->>IF: RefreshInterfaces()
  IF-->>API: NIC list (id, name, IPs)
  API-->>UI: JSON interfaces

  User->>UI: Start capture
  UI->>API: POST /capture/start/{id}
  API->>IF: GetByID(id) → pcap device name
  API->>CAP: StartCapture(pcap_id)
  Note over CAP: Spawns capture + parse + filter + reassembly goroutines
  API-->>UI: 200 OK

  loop Live feed
    CAP-->>WS: EventBus → Broadcaster → Hub
    WS-->>UI: JSON packet / stream events
  end

  User->>UI: Update filter
  UI->>API: POST /filter
  API->>CAP: UpdateFilter(rule)

  User->>UI: Stop capture
  UI->>API: POST /capture/stop
  API->>CAP: Stop() — cancel context, close pcap, wait workers
```

### Data pipeline (backend hot path)

When capture is running, each frame from the wire is duplicated into **two parallel pipelines**: per-packet parsing and TCP stream reassembly.

```mermaid
flowchart LR
  subgraph Capture["1. Capture"]
    H[pcap.Handle]
    ENG[capture.StartCapture]
    RAW[(rawChan)]
    H --> ENG --> RAW
  end

  subgraph Fanout["2. Fan-out"]
    RAW --> PARSE_IN[parseChan]
    RAW --> ASM_IN[assemblerChan]
  end

  subgraph PacketPath["3a. Packet path"]
    PARSE_IN --> PARSER[parser.StartParser]
    PARSER --> PARSED[(parsedChan)]
    PARSED --> FILTER[filter.StartFilter]
    FILTER --> FILTERED[(filteredChan)]
    FILTERED --> PUB_PKT[bus.Publish EventPacket]
  end

  subgraph StreamPath["3b. Stream path"]
    ASM_IN --> ASM[reassembly.Assembler]
    ASM --> SP[StreamParser]
    SP --> PUB_STR[bus.Publish EventStream]
  end

  subgraph Out["4. To browser"]
    PUB_PKT --> BUS[EventBus]
    PUB_STR --> BUS
    BUS --> BC[websocket.StartBroadcaster]
    BC --> HUB[Hub.broadcast]
    HUB --> CLIENT[Browser WebSocket]
  end
```

| Stage | Package / file | Role |
|-------|----------------|------|
| Open adapter | `internal/core/capture/pcap.go` | `pcap.OpenLive` + BPF filter |
| Read frames | `internal/core/capture/engine.go` | `PacketSource` → raw bytes |
| Fan-out | `internal/service/caputreservice.go` | Same raw frame to parser + assembler |
| Decode packet | `internal/core/parser/decoder.go` | IPv4, TCP/UDP, DNS fields → `models.Packet` |
| User filter | `internal/core/filter/` | `Manager` + rules from `POST /filter` |
| TCP streams | `internal/core/reassembly/` | Reassemble TCP → `StreamParser` (TLS SNI, etc.) |
| Events | `internal/core/bus/` | Pub/sub; slow consumers drop events |
| Realtime | `internal/api/websocket/` | JSON marshal → all connected clients |

### Backend startup wiring

`cmd/sniffer/main.go` calls `httpserver.StartServer()`, which:

1. Creates `InterfaceService` and `CaptureService`
2. Starts `websocket.Hub.Run()` in a goroutine
3. Subscribes the **Broadcaster** to `CaptureService`’s `EventBus`
4. Registers REST routes on `CentralHandler`
5. Attaches `GET /ws` for WebSocket upgrades

### Frontend design flow

```mermaid
flowchart TD
  APP[App.tsx]
  API[api.ts]
  HOOK[useSnifferSocket]
  TYPES[types.ts]

  APP -->|mount| API
  API -->|GET /interfaces| IFACES[Interface list state]
  APP -->|mount| HOOK
  HOOK -->|connect /ws| FEED[packets + streams state]

  APP -->|Start| API
  API -->|POST /capture/start/id| CAP_START[Backend starts pipeline]

  APP -->|Filter form| API
  API -->|POST /filter| CAP_FILTER[Backend filter manager]

  HOOK -->|onmessage| PARSE{Event type?}
  PARSE -->|packet| PKTS[Packet table max 400]
  PARSE -->|stream| STR[Stream table max 200]
  PARSE -->|handshake| HS[Connection status]

  APP --> TYPES
  HOOK --> TYPES
```

- **REST** (`api.ts`): interfaces, start/stop capture, filter — defaults to `http://localhost:8080` (or Vite proxy paths if you point `VITE_API_BASE` at the dev server).
- **WebSocket** (`useSnifferSocket.ts`): connects to `/ws` on the current host (Vite proxies to `:8080` in dev).
- **UI caps**: last 400 packets and 200 streams in memory to keep the dashboard responsive.

### Runtime connection (dev)

```
Browser (localhost:5173)
    │  HTTP  /interfaces, /capture, /filter  → Vite proxy → :8080
    │  WS    /ws                             → Vite proxy → :8080
    ▼
Go API (:8080)  →  libpcap  →  your NICs (Wi‑Fi, Ethernet, etc.)
```

---

## Project layout

```
PacketSniffer/
├── backend/
│   ├── cmd/sniffer/              # Entrypoint
│   └── internal/
│       ├── api/                  # HTTP, WebSocket, routes, handlers
│       ├── service/              # CaptureService, InterfaceService
│       ├── core/
│       │   ├── capture/          # pcap open + read loop
│       │   ├── parser/           # Per-packet decode (DNS, TCP, …)
│       │   ├── filter/           # Rule engine
│       │   ├── reassembly/       # TCP stream assembly
│       │   ├── bus/              # Event pub/sub
│       │   └── netfaces/         # List host interfaces
│       └── models/               # Shared structs
├── frontend/
│   ├── src/App.tsx               # Dashboard UI
│   ├── src/api.ts                # REST client
│   └── src/useSnifferSocket.ts   # Live feed
└── README.md
```

`backend/Dockerfile` and `frontend/Dockerfile` are optional (Linux images only). They are **not** recommended for capturing traffic on your host machine.

---

## API

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/interfaces` | List capture-capable interfaces |
| `POST` | `/capture/start/{id}` | Start capture on interface id |
| `POST` | `/capture/stop` | Stop capture |
| `POST` | `/filter` | Update filter rules (JSON) |
| WebSocket | `/ws` | Live packet and stream events |

Default BPF filter in code (`backend/internal/core/capture/pcap.go`):

`tcp or udp port 53 or tcp port 443`

---

## Troubleshooting

| Issue | What to try |
|-------|-------------|
| `go run` fails with CGO / pcap errors | Install Npcap (Windows) or `libpcap-dev` (Linux). Set `CGO_ENABLED=1`. Install `gcc`. |
| Empty interface list | Npcap not installed (Windows), or backend not run as admin / with sufficient rights. |
| Start capture fails | Run backend elevated (Windows: Administrator). Close other sniffers using the same adapter. |
| UI cannot connect | Backend must be on `:8080` first. Use `npm run dev` (not only `build`) so the Vite proxy is active. |
| WebSocket disconnected | Confirm backend is running; check browser devtools → Network → WS to `localhost:5173/ws`. |

---

## License

Add your license here if you publish the repo.
