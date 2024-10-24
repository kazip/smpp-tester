package main

import (
	"bufio"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/linxGnu/gosmpp"
	"github.com/linxGnu/gosmpp/pdu"
	"log"
	"os"
	"strings"
	"time"
)

func handlePDU() func(pdu.PDU) (pdu.PDU, bool) {
	return func(p pdu.PDU) (pdu.PDU, bool) {
		switch pd := p.(type) {
		case *pdu.Unbind:
			log.Println("Unbind Received")
			return pd.GetResponse(), true

		case *pdu.UnbindResp:
			log.Println("UnbindResp Received")

		case *pdu.SubmitSMResp:
			log.Println("SubmitSMResp Received")
			log.Println(pd)

		case *pdu.GenericNack:
			log.Println("GenericNack Received")

		case *pdu.EnquireLinkResp:
			log.Println("EnquireLinkResp Received")

		case *pdu.EnquireLink:
			log.Println("EnquireLink Received")
			return pd.GetResponse(), false

		case *pdu.DataSM:
			log.Println("DataSM Received")
			return pd.GetResponse(), false

		case *pdu.DeliverSM:
			log.Println("DeliverSM Received")
			log.Println(pd.Message.GetMessage())
			return pd.GetResponse(), false
		}
		return nil, false
	}
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

func main() {

	var opts struct {
		Rps           int    `long:"speed" short:"s" description:"rate per second" default:"50"`
		Host          string `long:"host" short:"H" description:"smpp server host" default:"localhost"`
		Port          uint   `long:"port" short:"P" description:"smpp server port" default:"2775"`
		SystemId      string `long:"system_id" short:"u" description:"SMPP systemId" required:"true"`
		Password      string `long:"password" short:"p" description:"SMPP password" required:"true"`
		SkipConfirm   bool   `long:"skip-confirm" short:"y"`
		Text          string `long:"text" short:"t" description:"SMS text" default:"load-test"`
		MaxCount      int    `long:"max-count" short:"m" description:"Maximum SMS number to send" default:"-1"`
		WaitDeliverSm int    `long:"wait-deliver-sm" short:"w" description:"Wait in seconds for deliver_sm after sending'" default:"10"`
		From          string `long:"from" short:"F" description:"source address" default:"test"`
		To            string `long:"to" short:"T" description:"destination address" default:"test"`
		TTL           int    `long:"ttl" description:"ttl" default:"60"`
	}

	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	log.Println(fmt.Sprintf("%s:%d", opts.Host, opts.Port))

	if !opts.SkipConfirm {
		c := askForConfirmation("Do you really want to send smpp traffic?")
		if !c {
			fmt.Println("Exiting...")
			os.Exit(1)
		}
	}

	auth := gosmpp.Auth{
		SMSC:       fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		SystemID:   opts.SystemId,
		Password:   opts.Password,
		SystemType: "",
	}

	trans, err := gosmpp.NewSession(
		gosmpp.TRXConnector(gosmpp.NonTLSDialer, auth),
		gosmpp.Settings{
			EnquireLink: 5 * time.Second,
			ReadTimeout: 10 * time.Second,
			OnAllPDU:    handlePDU(),
		},
		5*time.Second,
	)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer func() {
		_ = trans.Close()
	}()

	ticker := time.NewTicker(time.Second / time.Duration(opts.Rps))
	defer ticker.Stop()
	start := time.Now()
	n := 0
	for {
		<-ticker.C
		sendSubmitSm(trans, opts.Text, opts.From, opts.To, opts.TTL)
		n++
		currentRps := float64(n) / time.Since(start).Seconds()
		log.Println("Speed is: ", currentRps)
		if opts.MaxCount > 0 && n >= opts.MaxCount {
			break
		}
	}

	if opts.WaitDeliverSm > 0 {
		log.Println("Waiting for deliver_sms for", opts.WaitDeliverSm, "seconds")
		time.Sleep(time.Second * time.Duration(opts.WaitDeliverSm))
	}

}

func formatValidityPeriod(ttl int) string {
	// Создаем текущую метку времени
	now := time.Now().UTC()

	// Добавляем ttl к текущему времени
	validityTime := now.Add(time.Duration(ttl) * time.Second)

	// Форматируем строку по требуемому формату "yyMMddHHmmss000R"
	validityPeriod := validityTime.Format("060102150405") + "000+"

	return validityPeriod
}

func sendSubmitSm(trans *gosmpp.Session, text string, from string, to string, ttl int) {
	submitSm := pdu.NewSubmitSM().(*pdu.SubmitSM)

	srcAddr, _ := pdu.NewAddressWithAddr(from)
	dstAddr, _ := pdu.NewAddressWithAddr(to)

	message, err := pdu.NewShortMessage(text)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	submitSm.SourceAddr = srcAddr
	submitSm.DestAddr = dstAddr
	submitSm.Message = message
	submitSm.ValidityPeriod = formatValidityPeriod(ttl)

	err = trans.Transmitter().Submit(submitSm)

	if err != nil {
		log.Println(err)
	}
}
