package index

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeadingZeros(t *testing.T) {
	assert.Equal(t, Base2.LeadingZeros([]byte{byte(0b00001000)}), uint(4))
	assert.Equal(t, Base4.LeadingZeros([]byte{byte(0b00001000)}), uint(2))
	assert.Equal(t, Base8.LeadingZeros([]byte{byte(0b00001000)}), uint(1))
	assert.Equal(t, Base16.LeadingZeros([]byte{byte(0b00001000)}), uint(1))
	assert.Equal(t, Base32.LeadingZeros([]byte{byte(0b00001000)}), uint(0))

	assert.Equal(t, Base2.LeadingZeros([]byte{byte(0b00000000), byte(0b00010000)}), uint(11))
	assert.Equal(t, Base4.LeadingZeros([]byte{byte(0b00000000), byte(0b00010000)}), uint(5))
	assert.Equal(t, Base8.LeadingZeros([]byte{byte(0b00000000), byte(0b00010000)}), uint(3))
	assert.Equal(t, Base16.LeadingZeros([]byte{byte(0b00000000), byte(0b00010000)}), uint(2))
	assert.Equal(t, Base32.LeadingZeros([]byte{byte(0b00000000), byte(0b00010000)}), uint(2))
}
