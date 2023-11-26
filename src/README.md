Trust, semantic learning, and monitoring
========================================

These programs are proof of concept stubs and research tools associated with the project Trust semantic learning and monitoring with NLnet. Some of the code is duplicated to preserve the history of development. As the research and development progresses, some versions will become of mainly pedagogical interest, and only a few programs will be of interest to execute:

(DISCLAIMER: these programs worked at the time of writing, but may cease to work as dependencies fail to keep the same promises.)

 - `chatgpt_client.go`
 - `ngrams.go`
 - `ngrams-chinese.go`
 - `wikipedia_history.go` <verbose>
 - `wikipedia_history_db.go` <verbose>
 - `wikipedia_history_query.go` <topic>
 - `wikipedia_topic_query.go` <topic>

The files:

- pkg/TrustAndTrustability contains library code employing the Semantic Spacetime model

- src/

 - `chatgpt_client.go` - POC interface to chatGPT
 - `chinese-strokes.dat` - datafile containing a histogram/database of chinese characters, frequency and number of strokes
 - `gnuplot.in` - script file for generating the graphs in part 5 from a datafile produced by wikipedia_history.go
 - `html_client.go` - html, url reading stub
 - `http_client.go` - http protocol connector stub
 - `ngram+html.go` - merger between ngram code and html stub
 - `ngram-lib.go` - library refactored version of the stubs for reusability
 - `ngrams-chinese.go` - ngram summarization analysis for Chinese UTF8 text
 - `ngrams.go` - ngram summarization for Western alphabetic languages
 - `tcp_client.go` - tcp client stub to run together with tcp_server.go
 - `tcp_server.go` - tcp server stub to run together with tcp_client.go
 - `udp_client.go` - udp client stub to run together with upp_server.go
 - `udp_server.go` - udp server stub to run together with udp_client.go
 - `wikipedia_history.go` - html+ngram+wikipedia analysis, self contained output analysis generator
 - `wikipedia_history_db.go` - a database building version of the analysis developed in the previous
 - `wikipedia_history_ml.go` - a machine learning analysis of the wikipedia example using the API
 - `wikipedia_history_query.go` - prints out the episodic process history from a database once created
 - `wikipedia_topic_query.go` - prints out the sampled storyline for a topic from the database once created
 - `wikipedia_ml_query.go` - generate some graph output of the machine learning profles using the API

For specifics about how to run each program see the top notes in the source code.