
# The Trust Library API

This library has been developed as a proof of concept concerning the use of trust in on line applications. Unlike most discussions of trust, this is not a heuristic cryptotoken or credential scheme. Rather it develops the idea that trust is a running assessment of a relationship between observer and observed in a system. 

The code presented here is free to use, but it is to be thought of as a proof of concept, since much of it is tailored to the specific use cases for which data could be obtained. The code has been purposely designed with few entry points:

The purpose of any machine learning scheme is to condense inputs to provide a dimensional reduction of the original data.

## Data types and SST context

The library makes use of the Semantic Spacetime model context, which has supporting infrastructure currently based on ArangoDB.

A Go application using the library can open an Analytics context as follows:

```
    TT.InitializeSmartSpaceTime()

    var dbname string = "SST-ML"
    var dburl string = "http://localhost:8529" // ArangoDB port
    var user string = "root"
    var pwd string = "mark"

    G = TT.OpenAnalytics(dbname,dburl,user,pwd)
```


## Transaction wrappers

Two sets of functions for wrapping transactional events or critical sections parenthetically (with begin-end semantics).

* Functions that timestamp transactions at the moment of measurement according to the local system clock

```
 PromiseContext_Begin(g Analytics, name string) PromiseContext 
 PromiseContext_End(g Analytics, ctx PromiseContext) PromiseHistory 
```

* Functions with Golang time stamp supplied from outside, e.g. for offline analysis with epoch timestamps.

```
 StampedPromiseContext_Begin(g Analytics, name string, before time.Time) PromiseContext 
 StampedPromiseContext_End(g Analytics, ctx PromiseContext, after time.Time) PromiseHistory
```

## Periodic Key Value Storage

Most time based data can be classified according to their point in a week from Monday to Friday with finite resolution of five
minute intervals. Most system interactions have correlation times of up to 20 minutes. There is no need to collect nanosecond metrics for functioning systems. Each collection name (collname) represents a vector of values indexed by Unix or Golang timestamps
which are turned into 5 minute intervals.

These functions fail silently with empty data rather than raising internal errors. The main reason for failure is incorrect naming.

 SumWeeklyKV(g Analytics, collname string, t int64, value float64)
 LearnWeeklyKV(g Analytics, collname string, t int64, value float64)

 AddWeeklyKV_Unix(g Analytics, collname string, t int64, value float64)
 AddWeeklyKV_Go(g Analytics, collname string, t time.Time, value float64)


To retrieve data from the learning store ordered by the working week, 

```
 GetAllWeekMemory(g Analytics, collname string) []float64 
```

