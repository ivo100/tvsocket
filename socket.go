package tvsocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"strconv"
)

const (
	TradingViewSocketURL = "wss://data.tradingview.com/socket.io/websocket"
)

// Socket ...
type Socket struct {
	OnReceiveMarketDataCallback OnReceiveDataCallback
	OnErrorCallback             OnErrorCallback
	OnReceiveQuoteCallback      OnReceiveQuoteCallback
	conn                        *websocket.Conn
	isClosed                    bool
	quoteSessionID              string
	chartSessionID              string
	chartSessionName            string
}

// Connect - Connects and returns the trading view socket object
func Connect(
	onReceiveMarketDataCallback OnReceiveDataCallback,
	onErrorCallback OnErrorCallback,
	fields ...string,
) (socket SocketInterface, err error) {
	socket = &Socket{
		OnReceiveMarketDataCallback: onReceiveMarketDataCallback,
		OnErrorCallback:             onErrorCallback,
	}

	err = socket.Init(fields...)

	return
}

// Init connects to the tradingview web socket
func (s *Socket) Init(fields ...string) (err error) {
	s.isClosed = true
	s.chartSessionName = "price"
	s.quoteSessionID = s.generateSessionID(true)
	s.chartSessionID = s.generateSessionID(false)
	fmt.Printf("Session IDs: %s %s\n", s.quoteSessionID, s.chartSessionID)
	//fmt.Printf("Connecting to %s\n", TradingViewSocketURL)
	if s.conn, _, err = (&websocket.Dialer{}).Dial(TradingViewSocketURL, getHeaders()); err != nil {
		if s.OnErrorCallback != nil {
			s.onError(err, InitErrorContext)
		}
		return err
	}

	if err = s.checkFirstReceivedMessage(); err != nil {
		if s.OnErrorCallback != nil {
			s.onError(err, InitErrorContext)
		}
		return err
	}

	if err = s.sendConnectionSetupMessages(fields...); err != nil {
		if s.OnErrorCallback != nil {
			s.onError(err, InitErrorContext)
		}
		return err
	}

	s.isClosed = false
	go s.connectionLoop()

	return
}

// Close ...
func (s *Socket) Close() (err error) {
	if s.isClosed {
		return nil
	}
	s.isClosed = true
	return s.conn.Close()
}

// AddSymbol ...
func (s *Socket) AddSymbol(symbol string) (err error) {
	err = s.sendSocketMessage(
		//getSocketMessage("quote_add_symbols", []any{s.quoteSessionID, symbol, getFlags()}),
		getSocketMessage("quote_add_symbols", []any{s.quoteSessionID, symbol}),
	)
	return
}

// RemoveSymbol ...
func (s *Socket) RemoveSymbol(symbol string) (err error) {
	err = s.sendSocketMessage(
		getSocketMessage("quote_remove_symbols", []any{s.quoteSessionID, symbol}),
	)
	return
}

func (s *Socket) RequestQuotes(symbol string, bars int, interval string, onReceiveQuote OnReceiveQuoteCallback) (err error) {
	s.OnReceiveQuoteCallback = onReceiveQuote
	// 1. Add symbol
	//send_message(ws, "quote_add_symbols", [websocket_session, symbol, {"flags": ["force_permission"]}], )
	err = s.sendSocketMessage(
		getSocketMessage("quote_add_symbols", []any{
			s.quoteSessionID,
			symbol,
		}))
	if err != nil {
		return err
	}
	// 2. Resolve symbol
	m := getSocketMessage("resolve_symbol", []any{
		s.chartSessionID,
		"symbol_1",
		`={"symbol": "` + symbol + `"}`})
	err = s.sendSocketMessage(m)
	if err != nil {
		return err
	}

	// 3. Create series
	//time.Sleep(2 * time.Second)
	err = s.sendSocketMessage(
		getSocketMessage("create_series", []any{
			s.chartSessionID,
			s.chartSessionName,
			s.chartSessionName,
			"symbol_1",
			interval,
			bars,
		}))

	return err
}

func (s *Socket) checkFirstReceivedMessage() (err error) {
	var msg []byte
	//fmt.Printf("checkFirstReceivedMessage\n")
	_, msg, err = s.conn.ReadMessage()
	if err != nil {
		s.onError(err, ReadFirstMessageErrorContext)
		return
	}
	payload := msg[getPayloadStartingIndex(msg):]
	//fmt.Printf("payload %s\n", string(payload))
	var p map[string]any

	err = json.Unmarshal(payload, &p)
	if err != nil {
		s.onError(err, DecodeFirstMessageErrorContext)
		return
	}

	if p["session_id"] == nil {
		err = errors.New("cannot recognize the first received message after establishing the connection")
		s.onError(err, FirstMessageWithoutSessionIdErrorContext)
		return
	}

	return
}

func (s *Socket) generateSessionID(isQuoteSession bool) string {
	x := "cs_"
	if isQuoteSession {
		x = "qs_"
	}
	x += GetRandomString(12)
	return x
}

func (s *Socket) sendConnectionSetupMessages(fields ...string) (err error) {
	messages := []*SocketMessage{
		getSocketMessage("set_auth_token", []string{"unauthorized_user_token"}),
		getSocketMessage("chart_create_session", []string{s.chartSessionID, ""}),
		getSocketMessage("quote_create_session", []string{s.quoteSessionID}),
	}
	for _, msg := range messages {
		err = s.sendSocketMessage(msg)
		if err != nil {
			return
		}
	}
	// todo: separate method
	/*
	 */
	m := []string{s.quoteSessionID, "lp", "lp_time", "ch", "ch_time"}
	if len(fields) > 0 {
		for _, field := range fields {
			m = append(m, field)
		}
	}
	msg := getSocketMessage("quote_set_fields", m)
	//fmt.Printf("send %s\n", msg)
	_ = s.sendSocketMessage(msg)
	return
}

func (s *Socket) sendSocketMessage(p *SocketMessage) (err error) {
	payload, _ := json.Marshal(p)
	payloadWithHeader := "~m~" + strconv.Itoa(len(payload)) + "~m~" + string(payload)
	//fmt.Printf("Sending %s\n", payloadWithHeader)
	err = s.conn.WriteMessage(websocket.TextMessage, []byte(payloadWithHeader))
	if err != nil {
		s.onError(err, SendMessageErrorContext+" - "+payloadWithHeader)
		return
	}
	return
}

func (s *Socket) connectionLoop() {
	var readMsgError error
	var writeKeepAliveMsgError error

	for readMsgError == nil && writeKeepAliveMsgError == nil {
		if s.isClosed {
			break
		}

		var msgType int
		var msg []byte

		msgType, msg, readMsgError = s.conn.ReadMessage()
		//fmt.Printf("ReadMessage - Received msg type %d, payload: %s, err %v\n", msgType, string(msg), readMsgError)
		go func(msgType int, msg []byte) {
			if msgType != websocket.TextMessage {
				return
			}

			if isKeepAliveMsg(msg) {
				writeKeepAliveMsgError = s.conn.WriteMessage(msgType, msg)
				return
			}

			go s.parsePacket(msg)
		}(msgType, msg)
	}

	if readMsgError != nil {
		s.onError(readMsgError, ReadMessageErrorContext)
	}
	if writeKeepAliveMsgError != nil {
		s.onError(writeKeepAliveMsgError, SendKeepAliveMessageErrorContext)
	}
}

func (s *Socket) parsePacket(packet []byte) {
	var symbolsArr []string
	var dataArr []*QuoteData

	index := 0
	for index < len(packet) {
		payloadLength, err := getPayloadLength(packet[index:])
		if err != nil {
			fmt.Printf("Error while getting payload length - %v\n", err)
			s.onError(err, GetPayloadLengthErrorContext+" - "+string(packet))
			return
		}

		headerLength := 6 + len(strconv.Itoa(payloadLength))
		payload := packet[index+headerLength : index+headerLength+payloadLength]
		index = index + headerLength + len(payload)
		//fmt.Printf("> Payload %s\n", payload)

		symbol, data, err := s.parseJSON(payload)
		if err != nil {
			fmt.Printf("> parseJSON error %s\n", err.Error())
			continue
		}
		hloc, ok := data.([]HLOC)
		if ok {
			//fmt.Printf(">>> Received %s - %+v\n", symbol, x)
			if s.OnReceiveQuoteCallback != nil {
				s.OnReceiveQuoteCallback(symbol, hloc)
			}
			continue
		}
		//fmt.Printf(">>> Received %s - %+v\n", symbol, data)
		if data == nil {
			continue
		}
		dataArr = append(dataArr, data.(*QuoteData))
		symbolsArr = append(symbolsArr, symbol)
	}

	// TODO: fix this nested loop !!!

	for i := 0; i < len(dataArr); i++ {
		isDuplicate := false
		for j := i + 1; j < len(dataArr); j++ {
			if GetStringRepresentation(dataArr[i]) == GetStringRepresentation(dataArr[j]) {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			s.OnReceiveMarketDataCallback(symbolsArr[i], dataArr[i])
		}
	}
}

func (s *Socket) parseJSON(payload []byte) (symbol string, data any, err error) {
	var msg *SocketMessage

	err = json.Unmarshal(payload, &msg)
	if err != nil {
		s.onError(err, DecodeMessageErrorContext+" - "+string(payload))
		return
	}

	if msg.Message == "critical_error" || msg.Message == "error" {
		err = errors.New("Error -> " + string(payload))
		s.onError(err, DecodedMessageHasErrorPropertyErrorContext)
		return
	}

	if msg.Message == "timescale_update" {
		return parseTimeScaleUpdate(payload)
	}

	if msg.Message != "qsd" {
		//err = errors.New("ignored message (Not qsd), got: " + msg.Message)
		return
	}

	if msg.Payload == nil {
		err = errors.New("Msg does not include 'p' -> " + string(payload))
		s.onError(err, DecodedMessageDoesNotIncludePayloadErrorContext)
		return
	}
	p, isPOk := msg.Payload.([]any)
	if !isPOk {
		fmt.Printf("expected array\n")
		err = errors.New("There is something wrong with the payload - can't be parsed -> " + string(payload))
		fmt.Printf("err: %v\n", err)
		//s.onError(err, PayloadCantBeParsedErrorContext)
		return
	}

	if len(p) != 2 {
		fmt.Printf("expected array with len 2 got %d\n", len(p))
		err = errors.New("There is something wrong with the payload - can't be parsed -> " + string(payload))
		fmt.Printf("err: %v\n", err)
		//s.onError(err, PayloadCantBeParsedErrorContext)
		return
	}

	var decodedQuoteMessage *QuoteMessage
	err = mapstructure.Decode(p[1].(map[string]any), &decodedQuoteMessage)
	if err != nil {
		s.onError(err, FinalPayloadCantBeParsedErrorContext+" - "+string(payload))
		return
	}

	if decodedQuoteMessage.Status != "ok" || decodedQuoteMessage.Symbol == "" || decodedQuoteMessage.Data == nil {
		err = errors.New("There is something wrong with the payload - couldn't be parsed -> " + string(payload))
		s.onError(err, FinalPayloadHasMissingPropertiesErrorContext)
		return
	}
	symbol = decodedQuoteMessage.Symbol
	data = decodedQuoteMessage.Data
	return symbol, data, nil
}

func parseTimeScaleUpdate(payload []byte) (sym string, hloc []HLOC, err error) {
	d := make(map[string]any)
	err = json.Unmarshal(payload, &d)
	if err != nil {
		return
	}
	err = fmt.Errorf("parsing error")
	//fmt.Printf("d: %+v\n", d)
	p := d["p"].([]any)
	if len(p) < 2 {
		return
	}
	price := p[1]
	var amap map[string]any
	var a any
	if amap = price.(map[string]any); amap == nil {
		return
	}
	if a = amap["price"]; a == nil {
		return
	}
	if a = a.(map[string]any)["s"]; a == nil {
		return
	}

	hloc = make([]HLOC, 0)
	for _, v := range a.([]any) {
		amap = v.(map[string]any)
		//i := int((amap["i"]).(float64))
		vals := (amap["v"]).([]any)
		var h HLOC
		for j, val := range vals {
			switch j {
			case 0:
				h.Time = int64(val.(float64))
			case 1:
				h.Open = val.(float64)
			case 2:
				h.High = val.(float64)
			case 3:
				h.Low = val.(float64)
			case 4:
				h.Close = val.(float64)
			case 5:
				h.Volume = int64(val.(float64))
			default:
			}
		}
		//fmt.Printf("h: %+v\n", h)
		hloc = append(hloc, h)
	}
	return sym, hloc, nil
}

func (s *Socket) onError(err error, context string) {
	fmt.Printf("ONERROR Error: %v\n", err)
	if s.conn != nil {
		_ = s.conn.Close()
	}
	if s.OnErrorCallback != nil {
		s.OnErrorCallback(err, context)
	}
}

func getSocketMessage(m string, p any) *SocketMessage {
	return &SocketMessage{
		Message: m,
		Payload: p,
	}
}

func getFlags() *Flags {
	return &Flags{
		Flags: []string{"force_permission"},
	}
}

func isKeepAliveMsg(msg []byte) bool {
	return string(msg[getPayloadStartingIndex(msg)]) == "~"
}

func getPayloadStartingIndex(msg []byte) int {
	char := ""
	index := 3
	for char != "~" {
		char = string(msg[index])
		index++
	}
	index += 2
	return index
}

func getPayloadLength(msg []byte) (length int, err error) {
	char := ""
	index := 3
	lengthAsString := ""
	for char != "~" {
		char = string(msg[index])
		if char != "~" {
			lengthAsString += char
		}
		index++
	}
	length, err = strconv.Atoi(lengthAsString)
	return
}

func getHeaders() http.Header {
	headers := http.Header{}

	headers.Set("Accept-Encoding", "gzip, deflate, br")
	headers.Set("Accept-Language", "en-US,en;q=0.9,es;q=0.8")
	headers.Set("Cache-Control", "no-cache")
	headers.Set("Host", "data.tradingview.com")
	headers.Set("Origin", "https://www.tradingview.com")
	headers.Set("Pragma", "no-cache")
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.193 Safari/537.36")

	return headers
}
