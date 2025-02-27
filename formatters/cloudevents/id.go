package cloudevents

import (
	"fmt"

	"github.com/hashicorp/vault/sdk/helper/base62"
)

func newId() (string, error) {
	const op = "cloudevents.newId"
	id, err := base62.Random(10)
	if err != nil {
		return "", fmt.Errorf("%s: unable to generate id: %w", op, err)
	}
	return id, nil
}
