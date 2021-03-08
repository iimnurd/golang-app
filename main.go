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
     "github.com/opentracing/opentracing-go"
     "context"
    "io"
    "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
    "github.com/opentracing/opentracing-go/ext"
    

    // "github.com/opentracing/opentracing-go"
    // "github.com/opentracing/opentracing-go/ext"
    // "github.com/uber/jaeger-client-go"
    // jaegercfg "github.com/uber/jaeger-client-go/config"
    // jaegerlog "github.com/uber/jaeger-client-go/log"
    // "github.com/uber/jaeger-lib/metrics"
 
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






// init is invoked before main()
func init() {
    // loads values from .env into the system
    if err := godotenv.Load(); err != nil {
        log.Print("No .env file found")
    }
    
}



func main(){
   
    mux := http.NewServeMux()

   



    mux.HandleFunc("/webhook", echoHandler)
    mux.HandleFunc("/health", healthcheck)
   
    http.ListenAndServe(":8000", mux)
}








func (s Resp_time) MarshalJSON() ([]byte, error) {
    data := map[string]interface{}{
        key: s.GOLANG,
    }
    return json.Marshal(data)
}

func action(data Request, StartTime time.Time,  r *http.Request, tracer opentracing.Tracer) (jsonInBytes []byte){
    

    span := StartSpanFromRequest(tracer, r)
    defer span.Finish()

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
    

   return
  

}

func pop(alist *[]string) string {
    f:=len(*alist)
    rv:=(*alist)[f-1]
    *alist=append((*alist)[:f-1])
    return rv
 }


func forward(data Request, StartTime time.Time, r *http.Request, tracer opentracing.Tracer ) (newData []byte) {
  


url := pop(&data.Request)

datas := Request{
    data.Id,
    data.Request,
}
jsonInBytes, err:= json.Marshal(datas)
if err != nil {
    log.Fatalln(err) 
}




span := StartSpanFromRequest(tracer, r)
defer span.Finish()



ctx := opentracing.ContextWithSpan(context.Background(), span)


span2, _ := opentracing.StartSpanFromContext(ctx, "ping-send")
defer span2.Finish()

// if err_span != nil {
//     log.Fatalf("An Error Occured %v", err_span)
//  }

req, _ := http.NewRequest("POST", url, bytes.NewReader(jsonInBytes))

if err := Inject(span2, req); err != nil {
    return 
}
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
  
    thisServiceName := os.Getenv("SERVICE_NAME")
    tracer, closer := Init(thisServiceName)
    defer closer.Close()
    opentracing.SetGlobalTracer(tracer)

    start := time.Now()

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
        w.Write([]byte("OK"))


    
}
// A function to be wrapped
func slowFunc(s string, c chan string) {
    time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
    c <- "received " + s
   }

func Inject(span opentracing.Span, request *http.Request) error {
	return span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(request.Header))
}

// Extract extracts the inbound HTTP request to obtain the parent span's context to ensure
// correct propagation of span context throughout the trace.
func Extract(tracer opentracing.Tracer, r *http.Request) (opentracing.SpanContext, error) {
	return tracer.Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
}

func Init(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		ServiceName: service,

		// "const" sampler is a binary sampling strategy: 0=never sample, 1=always sample.
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},

		// Log the emitted spans to stdout.
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

func StartSpanFromRequest(tracer opentracing.Tracer, r *http.Request) opentracing.Span {
	spanCtx, _ := Extract(tracer, r)
	return tracer.StartSpan("request-receive", ext.RPCServerOption(spanCtx))
}
