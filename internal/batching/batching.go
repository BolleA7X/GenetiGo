// Package batching provides utility data structures and routines for
// splitting the population into batches, where each batch is processed
// by a separate goroutine.
package batching

// BatchInfo contains the start and end indeces of a batch. Indeces are
// referred to the original array or slice that contains all members of
// the population.
type BatchInfo struct {
	Start uint32 // Start index (inclusive)
	End   uint32 // End index (exclusive)
}

// BuildBatchesList returns a slice that tells how to split the given number of
// elements into the desired number of batches. It tries to split the elements
// evenly, but the last batch can be bigger than the others if the number of
// elements is not a multiple of the number of batches.
func BuildBatchesList(nElements uint32, nBatches uint32) []BatchInfo {
	if nElements == 0 {
		return []BatchInfo{}
	}

	var batches = make([]BatchInfo, 0, nBatches)
	var batchSize uint32 = nElements / nBatches

	var start uint32 = 0
	var end uint32 = batchSize
	for range nBatches {
		if end > nElements {
			end = nElements
		}
		batches = append(batches, BatchInfo{start, end})
		start = end
		end += batchSize
	}

	// Handle uneven divisions

	if batches[nBatches-1].End != nElements {
		batches[nBatches-1].End = nElements
	}

	return batches
}
