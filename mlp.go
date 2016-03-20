package opus

type MLP struct {
	layers int32,
	topo const *int32,
	weights const *int32
};

func mlp_process(m *MLP, in, out *float) {

	var j int32
	var hidden int16[MAX_NEURONS]
	var W := m.weights
	var W_index := 0

	for j = 0; j < m.topo[1]; j++ {
		var k int32
		var sum = W[W_index]
		W_index++
		
	}
	for j = 0; j < m.topo[2]; j++ {
	}
}
