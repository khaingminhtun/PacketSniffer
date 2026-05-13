package reassembly

import (
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/tcpassembly"
)

type Assembler struct {
	assembler *tcpassembly.Assembler
}

func New(factory *StreamFactory) *Assembler {
	pool := tcpassembly.NewStreamPool(factory)
	return &Assembler{
		assembler: tcpassembly.NewAssembler(pool),
	}
}

func (a *Assembler) Process(packet gopacket.Packet) {
	if packet.NetworkLayer() == nil {
		return
	}

	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer == nil {
		return
	}

	tcp := tcpLayer.(*layers.TCP)

	a.assembler.AssembleWithTimestamp(
		packet.NetworkLayer().NetworkFlow(),
		tcp,
		packet.Metadata().Timestamp,
	)
}

// Call periodically (important to avoid memory leak)
func (a *Assembler) FlushOlderThan(d time.Duration) {
	a.assembler.FlushOlderThan(time.Now().Add(-d))
}
