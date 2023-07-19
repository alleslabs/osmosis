package writer

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

type PubSubWriter struct {
}

func NewPubSubWriter(url string) *PubSubWriter {
	var opts []option.ClientOption

	opts = append(opts, option.WithCredentialsFile("cred.json"))
	client, err := pubsub.NewClient(context.Background(), "alles-playground", opts...)
	if err != nil {
		panic(err)
	}

	topic := client.Topic("projects/alles-playground/topics/chat-test-txs")

	res := topic.Publish(context.Background(), &pubsub.Message{
		Data: []byte("hello world"),
	})
	// The publish happens asynchronously.
	// Later, you can get the result from res:
	msgID, err := res.Get(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(msgID)

	return &PubSubWriter{}
}
