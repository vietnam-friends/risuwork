package fixture

import (
	"math/rand"
	"strings"

	"risuwork-benchmarker/scenario/model"

	"github.com/brianvoe/gofakeit/v6"
)

func init() {

	gofakeit.AddFuncLookup("industry_id", gofakeit.Info{
		Category:    "custom",
		Description: "Random industry id",
		Example:     "I01",
		Output:      "string",
		Generate: func(r *rand.Rand, m *gofakeit.MapParams, info *gofakeit.Info) (interface{}, error) {
			return Industries[r.Intn(len(Industries))].ID, nil
		},
	})

	gofakeit.AddFuncLookup("tags", gofakeit.Info{
		Category:    "custom",
		Description: "Random comma separated tags",
		Example:     "交通費支給,研修あり,資格取得支援制度",
		Output:      "string",
		Generate: func(r *rand.Rand, m *gofakeit.MapParams, info *gofakeit.Info) (interface{}, error) {
			return GenerateRandomTags(), nil
		},
	})
}

func GenerateCSUser() *model.CSUser {
	var cs model.CSUser
	gofakeit.Struct(&cs)
	return &cs
}

func GenerateCompany() *model.Company {
	var c model.Company
	gofakeit.Struct(&c)
	for _, industry := range Industries {
		if industry.ID == c.IndustryID {
			c.Industry = industry.Name
			break
		}
	}
	return &c
}

func GenerateCLUser() *model.CLUser {
	var cl model.CLUser
	gofakeit.Struct(&cl)
	return &cl
}

func GenerateJob() *model.Job {
	var j model.Job
	gofakeit.Struct(&j)
	return &j
}

// 52種類(a-zA-Z) 32 文字
// uuid よりも衝突確率が低い
func GenerateRandomString() string {
	return gofakeit.LetterN(32)
}

// GenerateRandomTags タグリストを生成する
func GenerateRandomTags() string {
	return GenerateRandomTagsN(rand.Intn(3) + 1) // 1 to 3 tags
}

// GenerateRandomTagsN 指定の個数のタグリストを生成する
func GenerateRandomTagsN(numTags int) string {
	// 被らないようにケア
	var tags []string
	selectedTags := make(map[int]bool)
	for len(selectedTags) < numTags {
		i := rand.Intn(len(TagsList))
		if !selectedTags[i] {
			selectedTags[i] = true
			tags = append(tags, TagsList[i])
		}
	}
	return strings.Join(tags, ",")
}

// GenerateRandomEmail 適当なEmailアドレスを生成する
func GenerateRandomEmail() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b) + "@example.com"
}

func GenerateRandomPassword() (string, string) {
	p := Passwords[rand.Intn(len(Passwords))]
	return p.Raw, p.Hash
}

func GenerateJobDescription() (string, string) {
	p := JobDescriptions[rand.Intn(len(JobDescriptions))]
	return p.Title, p.Description
}

func GenerateJobSalary() int {
	return (rand.Intn(10) + 5) * 1000000
}

// RandInt returns int num in the closed interval [min,max]
func RandInt(min int, max int) int {
	return min + 1 + rand.Intn(max-min)
}
