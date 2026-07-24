package mapper

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/sarbojitrana/pii-redactor/internal/detect"
)

type Mapper struct{}

func NewMapper() *Mapper {
	return &Mapper{}
}

func (m *Mapper) Map(match detect.Match) string {
	seed := hashSeed(string(match.Category), match.Value)
	switch match.Category {
	case detect.CategoryEmail:
		return fakeEmail(seed)
	case detect.CategoryPhone:
		return fakePhone(seed)
	case detect.CategorySSN:
		return fakeSSN(seed)
	case detect.CategoryCreditCard:
		return fakeCreditCard(seed)
	case detect.CategoryIP:
		return fakeIP(seed)
	case detect.CategoryDOB:
		return fakeDOB(seed)
	case detect.CategoryAddress:
		return fakeAddress(seed)
	case detect.CategoryName:
		return fakeName(seed)
	case detect.CategoryCompany:
		return fakeCompany(seed)
	case detect.CategoryCIN:
		return fakeCIN(seed)
	case detect.CategoryPAN:
		return fakePAN(seed)
	case detect.CategoryISIN:
		return fakeISIN(seed)
	default:
		return "[REDACTED]"
	}
}

func hashSeed(category, value string) uint64 {
	sum := sha256.Sum256([]byte(category + "|" + value))
	return binary.BigEndian.Uint64(sum[:8])
}

func newRand(seed uint64) *rand.Rand {
	return rand.New(rand.NewSource(int64(seed)))
}

func randomAlphaLower(r *rand.Rand, n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func fakeEmail(seed uint64) string {
	r := newRand(seed)
	local := randomAlphaLower(r, 8)
	return fmt.Sprintf("%s@example.com", local)
}

func fakePhone(seed uint64) string {
	r := newRand(seed)
	first := 6 + r.Intn(4)
	rest := r.Intn(1000000000)
	return fmt.Sprintf("%d%09d", first, rest)
}

func fakeSSN(seed uint64) string {
	r := newRand(seed)
	area := 900 + r.Intn(100)
	group := r.Intn(100)
	serial := r.Intn(10000)
	return fmt.Sprintf("%03d-%02d-%04d", area, group, serial)
}

func luhnCheckDigit(digits []int) int {
	sum := 0
	parity := len(digits) % 2
	for i, d := range digits {
		if i%2 == parity {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	return (10 - (sum % 10)) % 10
}

func fakeCreditCard(seed uint64) string {
	r := newRand(seed)
	digits := make([]int, 15)
	digits[0], digits[1], digits[2], digits[3] = 4, 0, 0, 0
	for i := 4; i < 15; i++ {
		digits[i] = r.Intn(10)
	}
	check := luhnCheckDigit(digits)
	full := append(digits, check)
	var b strings.Builder
	for i, d := range full {
		if i > 0 && i%4 == 0 {
			b.WriteByte(' ')
		}
		b.WriteString(strconv.Itoa(d))
	}
	return b.String()
}

func fakeIP(seed uint64) string {
	r := newRand(seed)
	return fmt.Sprintf("192.0.2.%d", 1+r.Intn(254))
}

func fakeDOB(seed uint64) string {
	r := newRand(seed)
	year := 1960 + r.Intn(46)
	month := 1 + r.Intn(12)
	day := 1 + r.Intn(28)
	return fmt.Sprintf("%02d/%02d/%04d", day, month, year)
}

func fakeAddress(seed uint64) string {
	r := newRand(seed)
	plotNo := 1 + r.Intn(999)
	sector := 1 + r.Intn(30)
	return fmt.Sprintf("Plot No. %d, Example Sector %d, Sample Town, Example State - 000000", plotNo, sector)
}

var fakeFirstNames = []string{"Aarav", "Vivaan", "Aditya", "Ananya", "Diya", "Ishaan", "Kavya", "Meera", "Rohan", "Saanvi"}
var fakeLastNames = []string{"Sharma", "Verma", "Iyer", "Nair", "Reddy", "Gupta", "Kapoor", "Menon", "Rao", "Bose"}

func fakeName(seed uint64) string {
	r := newRand(seed)
	first := fakeFirstNames[r.Intn(len(fakeFirstNames))]
	last := fakeLastNames[r.Intn(len(fakeLastNames))]
	return first + " " + last
}

var fakeCompanySuffixes = []string{"Industries", "Enterprises", "Ventures", "Holdings", "Systems", "Traders"}

func fakeCompany(seed uint64) string {
	r := newRand(seed)
	n := 1 + r.Intn(9999)
	suffix := fakeCompanySuffixes[r.Intn(len(fakeCompanySuffixes))]
	return fmt.Sprintf("Sample %s %d Private Limited", suffix, n)
}

func fakeCIN(seed uint64) string {
	r := newRand(seed)
	industry := r.Intn(100000)
	year := 1980 + r.Intn(45)
	serial := r.Intn(1000000)
	return fmt.Sprintf("U%05dZZ%04dPLC%06d", industry, year, serial)
}

func fakePAN(seed uint64) string {
	r := newRand(seed)
	n := r.Intn(10000)
	last := 'A' + rune(r.Intn(26))
	return fmt.Sprintf("ZZZZZ%04d%c", n, last)
}

func fakeISIN(seed uint64) string {
	r := newRand(seed)
	var b strings.Builder
	b.WriteString("ZZ")
	for i := 0; i < 9; i++ {
		if r.Intn(3) == 0 {
			b.WriteByte(byte('A' + r.Intn(26)))
		} else {
			b.WriteByte(byte('0' + r.Intn(10)))
		}
	}
	b.WriteByte(byte('0' + r.Intn(10)))
	return b.String()
}