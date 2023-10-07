package mail

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/web3dev6/simplebank/util"
)

func TestSendEmailWithGamil(t *testing.T) {
	// load config from app.env
	config, err := util.LoadConfig("../.")
	require.NoError(t, err)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	subject := "Simple Bank Test Email"
	content := `
    <h1>Simple Bank</h1>
    <p>This is a test message from Simple Bank Golang Project for a "welcome user" email</p> 
    <p>Checkout the cool dev behind this @<a href= "https://github.com/web3dev6">Sarthak Joshi</a></p>
    `
	to := []string{"sarthakjoshi.in@gmail.com"}
	attachFiles := []string{"../simple-bank.pdf"}

	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}
