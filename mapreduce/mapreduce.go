// A simple map/reduce worker.
package mapreduce


// The map function
type MapFn func (item interface{}) interface{}

// The reduce function.  This is given a channel to the results of the map function. When
// all the map functions have finished work, the channel will be closed.
type ReduceFn func (items chan interface{})

type SimpleMapReduce struct {
    mappers             int
    hasStarted          bool
    mapFn               MapFn
    reduceFn            ReduceFn
    workQueue           chan interface{}
    reduceQueue         chan interface{}
    mappersFinished     []chan bool
    reducedFinished     chan bool
}

// Creates a new worker.
func NewSimpleMapReduce(mappers int, mapQueueSize int, reduceQueueSize int) *SimpleMapReduce {
    return &SimpleMapReduce{
        mappers:         mappers,
        hasStarted:      false,
        mapFn:           func (item interface{}) interface{} {
            return item
        },
        reduceFn:        nil,
        workQueue:       make(chan interface{}, mapQueueSize),
        reduceQueue:     make(chan interface{}, reduceQueueSize),
        mappersFinished: make([]chan bool, mappers),
        reducedFinished: make(chan bool),
    }
}

// Sets the mapper function
func (w *SimpleMapReduce) Map(mapFn MapFn) *SimpleMapReduce {
    w.mapFn = mapFn
    return w
}

// Sets the reduction function
func (w *SimpleMapReduce) Reduce (reduceFn ReduceFn) *SimpleMapReduce {
    w.reduceFn = reduceFn
    return w
}

// Start the map reducer.
func (w *SimpleMapReduce) Start() *SimpleMapReduce {
    if (w.hasStarted) {
        return w
    }

    w.hasStarted = true

    for i := 0; i < w.mappers; i++ {
        mapFn := w.mapFn
        mapperFinished := make(chan bool)
        w.mappersFinished[i] = mapperFinished

        // Parallel function which performs the map and adds the result to the reduction queue
        go func() {
            for item := range w.workQueue {
                res := mapFn(item)
                w.reduceQueue <- res
            }
            close(mapperFinished)
        }()
    }

    // If a reduction function is specified, start it.  Otherwise, simply close the reducedFinish
    // channel.
    if (w.reduceFn != nil) {
        go func() {
            w.reduceFn(w.reduceQueue)
            close(w.reducedFinished)
        }()
    } else {
        close(w.reducedFinished)
    }

    return w
}

// Add items to the queue.
func (w *SimpleMapReduce) Push(item interface{}) {
    w.workQueue <- item
}

// Shuts down the queue and wait for the reducer to finish
func (w *SimpleMapReduce) Close() {
    close(w.workQueue)
    for _, f := range w.mappersFinished {
        <-f
    }

    close(w.reduceQueue)
    <-w.reducedFinished
}

