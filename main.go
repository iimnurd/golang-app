package main

import(
    "net/http"
    "encoding/json"
    "math/rand"
    "time"
    "fmt"
    "log"
    "io/ioutil"
    "bytes"
    "github.com/joho/godotenv"
    "os"
    "strconv"
    
    //  "context"
  
   

    "github.com/prometheus/client_golang/prometheus"
     "github.com/prometheus/client_golang/prometheus/promhttp"

    // "github.com/opentracing/opentracing-go"
    // "github.com/opentracing/opentracing-go/ext"
    // "github.com/uber/jaeger-client-go"
    // jaegercfg "github.com/uber/jaeger-client-go/config"
    // jaegerlog "github.com/uber/jaeger-client-go/log"
    // "github.com/uber/jaeger-lib/metrics"
    "go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/trace"
    "go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	// "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
    // "go.opentelemetry.io/otel/semconv"

    "go.opentelemetry.io/otel/baggage"
    "net/http/httptrace"
    "go.opentelemetry.io/otel/propagation"

 
)
var key string = ""

type StartTime time.Time
type EndTime float64

//User defines model for storing account details in database
type Request struct {
    Id string `json:"id"`
    Request []string `json:"request"`
   // CreatedAt time.Time
}

type Resp_time struct {
    GOLANG float64 

  }
  
  type Response struct {
    Id string `json:"id"`
    Number  int `json:"number"`
    Response_time Resp_time `json:"response_time"`
  }
  type Combined struct {
    Response_time []string `json:"response_time"`
}


// create a new counter vector
var getForwardCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_get_forward_count", // metric name
		Help: "Number of forwward request.",
	},
	[]string{"status"}, // labels
)

// create a new counter vector
var getActionCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_get_action_count", // metric name
		Help: "Number of action request.",
	},
	[]string{"status"}, // labels
)

var getHealthCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_get_health_count", // metric name
		Help: "Number of health request.",
	},
	[]string{"status"}, // labels
)

var getActionLatency = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_get_action_duration_seconds",
		Help:    "Latency of action method request in second.",
		Buckets: prometheus.LinearBuckets(0.01, 0.05, 10),
	},
	[]string{"status"},
)

var getForwardLatency = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_get_forward_duration_seconds",
		Help:    "Latency of forward method request in second.",
		Buckets: prometheus.LinearBuckets(0.01, 0.05, 10),
	},
	[]string{"status"},
)

var getHealthLatency = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_get_health_duration_seconds",
		Help:    "Latency of get_health request in second.",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	},
	[]string{"status"},
)




// init is invoked before main()
func init() {
    // loads values from .env into the system
    if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }

  
      
        prometheus.MustRegister(getForwardCounter)
        prometheus.MustRegister(getActionCounter)
        prometheus.MustRegister(getHealthCounter)
        prometheus.MustRegister(getForwardLatency)
        prometheus.MustRegister(getActionLatency)
        prometheus.MustRegister(getHealthLatency)
    
    
}



func main(){
   
    mux := http.NewServeMux()

   



    mux.HandleFunc("/webhook", echoHandler)
    mux.HandleFunc("/health", healthcheck)
    mux.Handle("/metrics", promhttp.Handler())
    http.ListenAndServe(":8000", mux)

    // http.Handle("/metrics", promhttp.Handler())
    // http.Handle("/webhook", echoHandler)

	// println("listening..")
	// http.ListenAndServe(":8000", nil)
}








func (s Resp_time) MarshalJSON() ([]byte, error) {
    data := map[string]interface{}{
        key: s.GOLANG,
    }
    return json.Marshal(data)
}

func action(data Request, StartTime time.Time,  r *http.Request, tracer trace.Tracer) (jsonInBytes []byte){
    
    var status string
	defer func() {
        // increment the counter on defer func
		getActionCounter.WithLabelValues(status).Inc()
	}()
    timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		getActionLatency.WithLabelValues(status).Observe(v)
        }))
        defer func() {
            timer.ObserveDuration()
        }()
        
    
    uk := attribute.Key("username")
    ctx := r.Context()
		span := trace.SpanFromContext(ctx)
		username := baggage.Value(ctx, uk)
		span.AddEvent("handling this...", trace.WithAttributes(uk.String(username.AsString())))

	


    currentTime := time.Now()

    diff := currentTime.Sub(StartTime)

    datas := Response{
        data.Id,
        rand.Intn(1000), 
        Resp_time{
            GOLANG: diff.Seconds()*1000, //seconds to milisecond

        },
    }
    if (os.Getenv("DEBUG") == "true"){
        key = os.Getenv("APP_NAME")+"-action-"+strconv.Itoa(rand.Intn(1000))
        }else {
        key = os.Getenv("APP_NAME")
        }
    jsonInBytes, _ = json.Marshal(datas)
    status = "success"

   return
  

}

func pop(alist *[]string) string {
    f:=len(*alist)
    rv:=(*alist)[f-1]
    *alist=append((*alist)[:f-1])
    return rv
 }


func forward(data Request, StartTime time.Time, r *http.Request, tracer trace.Tracer ) (newData []byte) {
    var status string
	defer func() {
        // increment the counter on defer func
		getForwardCounter.WithLabelValues(status).Inc()
	}()

    timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		getForwardLatency.WithLabelValues(status).Observe(v)
        }))
        defer func() {
            timer.ObserveDuration()
        }()
url := pop(&data.Request)

datas := Request{
    data.Id,
    data.Request,
}
jsonInBytes, err:= json.Marshal(datas)
if err != nil {
    log.Fatalln(err) 
}



    uk := attribute.Key("username")
ctx := r.Context()
    span := trace.SpanFromContext(ctx)
    username := baggage.Value(ctx, uk)
    span.AddEvent("handling this...", trace.WithAttributes(uk.String(username.AsString())))

    

ctx = httptrace.WithClientTrace(ctx, otelhttptrace.NewClientTrace(ctx))


req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonInBytes))

resp, err := http.DefaultClient.Do(req)

   if err != nil {
      log.Fatalf("An Error Occured %v", err)
   }
   defer resp.Body.Close()
//Read the response body
   body, err := ioutil.ReadAll(resp.Body)
   if err != nil {
      log.Fatalln(err)
   }
   currentTime := time.Now()

   diff := currentTime.Sub(StartTime)
   
    var m,n map[string]interface{}
    err2 := json.Unmarshal(body, &m)
   
    resp_data := Resp_time{
        GOLANG: diff.Seconds()*1000,  //seconds to milisecond
        
    }
    
    if (os.Getenv("DEBUG") == "true"){
    key = os.Getenv("APP_NAME")+"-forward-"+strconv.Itoa(rand.Intn(1000))
    }else {
    key = os.Getenv("APP_NAME")
    }
    jb, _ := json.Marshal(resp_data)
    json.Unmarshal(jb, &n)

    jb2, _ := json.Marshal(m["response_time"])
    json.Unmarshal(jb2, &n)

    
    m["response_time"] = n

    newData, err2 = json.Marshal(m)
    if err2 != nil {
        log.Fatalln(err2)
     }
    
     status = "success"
    
    return 
}


func getResponse(body []byte) (*Response, error) {
    var s = new(Response)
    err := json.Unmarshal(body, &s)
    if(err != nil){
        fmt.Println("whoops:", err)
    }
    return s, err
}


func echoHandler(w http.ResponseWriter, r *http.Request){
  
   

    start := time.Now()
    flush := initTracer()
	defer flush()

    tracer := otel.Tracer("component-main")
  
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
    

    request := Request{} //initialize empty user
    
    //Parse json request body and use it to set fields on user
    //Note that user is passed as a pointer variable so that it's fields can be modified
    err := json.NewDecoder(r.Body).Decode(&request)
    if err != nil{
        panic(err)
    }

    
    
    if (len(request.Request)<= 1){
        b := action(request, start, r, tracer)
        
    w.Header().Set("Content-Type","application/json")
    
    w.Write(b)
    

    }else {
        b := forward(request, start, r, tracer)
        w.Header().Set("Content-Type","application/json")
    
        w.Write(b)
    

    }
}

    func healthcheck(w http.ResponseWriter, r *http.Request){
        var status string
        defer func() {
            // increment the counter on defer func
            getHealthCounter.WithLabelValues(status).Inc()
        }()


	    timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		getHealthLatency.WithLabelValues(status).Observe(v)
        }))
        defer func() {
            timer.ObserveDuration()
        }()



        status="success"
        w.Write([]byte("OK"))


    
}
// A function to be wrapped
func slowFunc(s string, c chan string) {
    time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
    c <- "received " + s
   }



// initTracer creates a new trace provider instance and registers it as global trace provider.
func initTracer() func() {
	// Create and install Jaeger export pipeline.
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint("http://jaeger-collector-httttp:14268/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "trace-demo",
			Tags: []attribute.KeyValue{
				attribute.String("exporter", "jaeger"),
				attribute.Float64("float", 312.23),
			},
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		log.Fatal(err)
	}
    
	return flush
}



// prometheus : https://www.jajaldoang.com/post/monitor-golang-app-with-prometheus-grafana/
// https://stackoverflow.com/questions/47138461/get-total-requests-in-a-period-of-time
// https://www.gwos.com/application-monitoring-with-the-prometheus-client-and-groundwork-monitor/
