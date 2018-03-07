package statsbot

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	statsDB *gorm.DB
)

type Category struct {
	ID          int    `gorm:"AUTO_INCREMENT" gorm:"primary_key"`
	Name        string `gorm:"size:25;unique;index"`
	FullName    string `gorm:"size:25"`
	Min         int    `gorm:"DEFAULT:0`
	Max         int    `gorm:"DEFAULT:100000`
	Order       int    `gorm:"DEFAULT:0"`
	OptionValue bool   `gorm:"DEFAULT:false"`
}

type User struct {
	ID        int    `gorm:"primary_key"`
	DiscordID string `gorm:"size:20;unique;index"`
	Name      string `gorm:"size:20;index"`
	Admin     bool   `gorm:DEFAULT:false"`
}

type Stat struct {
	gorm.Model
	Category      Category
	CategoryID    int `gorm:"unique_index:user_stat"`
	User          User
	UserID        int `gorm:"unique_index:user_stat" sql:"type:bigint REFERENCES user(id)"`
	Value         int
	OptionalValue string `gorm:"DEFAULT:NULL"`
	Verified      bool   `gorm:"DEFAULT:true"`
}

var categories = []Category{
	{Name: "jogger", FullName: "Jogger", Min: 0, Max: 50000, Order: 1},
	{Name: "collector", FullName: "Collector", Min: 0, Max: 500000, Order: 2},
	{Name: "scientist", FullName: "Scientist", Min: 0, Max: 50000, Order: 3},
	{Name: "breeder", FullName: "Breeder", Min: 0, Max: 50000, Order: 4},
	{Name: "backpacker", FullName: "Backpacker", Min: 0, Max: 500000, Order: 5},
	{Name: "battlegirl", FullName: "Battle Girl", Min: 0, Max: 50000, Order: 6},
	{Name: "berrymaster", FullName: "Berry Master", Min: 0, Max: 500000, Order: 7},
	{Name: "gymleader", FullName: "Gym Leader", Min: 0, Max: 500000, Order: 8},
	{Name: "raidchampion", FullName: "Raid Champion", Min: 0, Max: 10000, Order: 9},
	{Name: "legendaryraid", FullName: "Legendary Raid", Min: 0, Max: 10000, Order: 10},
	{Name: "goldgyms", FullName: "Gold Gym Badges", Min: 0, Max: 50000, Order: 11},
	{Name: "totalxp", FullName: "Total XP", Min: 0, Max: 100000000, Order: 12},
}

var admins = []User{
	{DiscordID: "221247558008307713", Name: "Haynes", Admin: true},
	{DiscordID: "223097664961511426", Name: "Kateri315", Admin: true},
	{DiscordID: "162112652691111937", Name: "Alletzhauser", Admin: true},
}

func InitDB() (err error) {
	username := "statsuser"
	password := "statspass"

	statsDB, err = gorm.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True", username, password, "stats"))
	if err != nil {
		return err
	}

	err = statsDB.DB().Ping()
	if err != nil {
		return err
	}

	err = createDatabase()

	return err
}

func createDatabase() (err error) {
	log.Println("Creating tables...")
	statsDB.AutoMigrate(&Category{}, &User{}, &Stat{})
	/*statsDB.CreateTable(&Category{})
	for _, category := range categories {
		statsDB.Create(&category)
	}
	log.Println("Creating User table...")
	statsDB.CreateTable(&User{})
	for _, user := range admins {
		statsDB.Create(&user)
	}
	log.Println("Creating Stat table...")
	statsDB.CreateTable(&Stat{})
	*/

	return nil
}

func (Category) TableName() string {
	return "Category"
}

func GetCategory(s string) (category Category, err error) {
	res := statsDB.Where("name = ?", s).First(&category)
	fmt.Println("Got %+v\n", category)

	return category, res.Error
}

func GetCategories() ([]Category, error) {
	var categories []Category
	res := statsDB.Order("name asc").Find(&categories)

	return categories, res.Error
}

func PrintCategories() (string, error) {
	var categories []Category
	res := statsDB.Order("name asc").Find(&categories)
	if res.Error != nil {
		return "", res.Error
	}

	message := "Current Categories:\n"
	for _, cat := range categories {
		message += cat.FullName + " (" + cat.Name + ")" + "\n"
	}

	return message, nil
}

func (c *Category) Validate(value int) bool {
	if value < c.Min {
		return false
	} else if value > c.Max {
		return false
	}

	return true
}

func (c Category) AddStat(u User, v int) error {
	if !c.Validate(v) {
		return ERR_INVALID_VALUE
	}

	err := NewStat(c, u, v)
	if err != nil {
		return err
	}
	return nil
}

func (c *Category) GetAll() (stats []Stat, err error) {
	res := statsDB.Model(&c).Preload("User").Order("value desc").Related(&stats)

	return stats, res.Error
}

func (c *Category) PrintStats() (string, error) {
	var message string

	stats, err := c.GetAll()
	if err != nil {
		return message, err
	}

	rank := 1
	message = c.FullName + "\n"
	for _, stat := range stats {
		message += fmt.Sprintf("%d. %s %d\n", rank, stat.User.Name, stat.Value)
		rank++
	}

	return message, err
}

func (User) TableName() string {
	return "User"
}

func (u *User) Insert() error {
	res := statsDB.Where("discord_id = ?", u.DiscordID).Assign(User{Name: u.Name}).FirstOrCreate(&u)

	return res.Error
}

func PrintUsers() (string, error) {
	var users []User
	res := statsDB.Order("name asc").Find(&users)
	if res.Error != nil {
		return "", res.Error
	}

	message := "Current Stats Users:\n"
	for _, user := range users {
		message += user.Name + "\n"
	}

	return message, nil
}

func GetUser(s string) (user User, err error) {
	res := statsDB.Where("name = ? OR discord_id = ?", s, s).First(&user)

	return user, res.Error
}

func CheckAdmin(s string) bool {
	var user User
	_ = statsDB.Where("discord_id = ?", s).First(&user)

	return user.Admin
}

func (Stat) TableName() string {
	return "Stat"
}

func NewStat(c Category, u User, v int) error {
	stat := Stat{
		Category: c,
		User:     u,
	}
	fmt.Println(stat)
	res := statsDB.Where(Stat{CategoryID: c.ID, UserID: u.ID}).Assign(Stat{Value: v}).FirstOrCreate(&stat)
	if res.Error != nil {
		log.Printf("Error create new stat: %+v\n", res.Error)
		return res.Error
	}
	return nil
}
