package main

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
)

type record struct {
	gorm.Model
	Repository string    `gorm:"index;unique" json:"repo"`
	Version    string    `gorm:"index" json:"version"`
	RebootTime time.Time `gorm:"index" json:"time"`
	Ports      string    `json:"ports"` //5000,8081
	Env        string    `json:"env"`
	Vols       string    `json:"vols"`
}

var db *gorm.DB

func initdb() {
	var err error
	db, err = gorm.Open("sqlite3", "/usr/local/var/sirdm.db")
	if err != nil {
		log.Fatal("db init fail!")
	}
	db.AutoMigrate(&record{})
}

func closedb() {
	db.Close()
}

func queryRecords(rs *[]record) error {
	return db.Find(rs).Error
}

func updateRecord(repo string, r record) *record {
	rr := record{}
	db.FirstOrCreate(&rr, record{Repository: repo})
	db.Model(rr).Update(r)
	return &rr
}

func getRecord(r *record, repo string) error {
	r.Repository = repo
	return db.FirstOrCreate(r, record{Repository: repo}).Error
}
