# GORM

The fantastic ORM library for Golang, aims to be developer friendly.

WARNING: mysql 8.0.19 is the last successfully tested, begging from 8.0.20 tests fail

[![go report card](https://goreportcard.com/badge/github.com/jinzhu/gorm "go report card")](https://goreportcard.com/report/github.com/jinzhu/gorm)
[![wercker status](https://app.wercker.com/status/8596cace912c9947dd9c8542ecc8cb8b/s/master "wercker status")](https://app.wercker.com/project/byKey/8596cace912c9947dd9c8542ecc8cb8b)
[![codecov](https://codecov.io/gh/jinzhu/gorm/branch/master/graph/badge.svg)](https://codecov.io/gh/jinzhu/gorm)
[![Join the chat at https://gitter.im/jinzhu/gorm](https://img.shields.io/gitter/room/jinzhu/gorm.svg)](https://gitter.im/jinzhu/gorm?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Open Collective Backer](https://opencollective.com/gorm/tiers/backer/badge.svg?label=backer&color=brightgreen "Open Collective Backer")](https://opencollective.com/gorm)
[![Open Collective Sponsor](https://opencollective.com/gorm/tiers/sponsor/badge.svg?label=sponsor&color=brightgreen "Open Collective Sponsor")](https://opencollective.com/gorm)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/jinzhu/gorm?status.svg)](https://godoc.org/github.com/jinzhu/gorm)

## Overview

* Full-Featured ORM (almost)
* Associations (Has One, Has Many, Belongs To, Many To Many, Polymorphism)
* Hooks (Before/After Create/Save/Update/Delete/Find)
* Preloading (eager loading)
* Transactions
* Composite Primary Key
* SQL Builder
* Auto Migrations
* Logger
* Extendable, write Plugins based on GORM callbacks
* Every feature comes with tests
* Developer Friendly

## Getting Started

* GORM Guides [https://gorm.io](https://gorm.io)

## SQL-injection prevention
We made some security patch in order to avoid SQL-injections. Changes implies that you are must use prepared statement
when calling the following methods:
```golang
	// these methods should be called using prepared statement
	DB.Delete(&item, "name = 'Foo'") 		// -> DB.Delete(&item, "name = ?", "Foo")
	DB.Find(&item, "name = 'Foo'") 			// -> DB.Find(&item, "name = ?", "Foo")
	DB.First(&item, "name = 'Foo'") 		// -> DB.First(&item, "name = ?", "Foo")
	DB.FirstOrCreate(&item, "name = 'Foo'")	// -> DB.FirstOrCreate(&item, "name = ?", "Foo")
	DB.FirstOrInit(&item, "name = 'Foo'")	// -> DB.FirstOrInit(&item, "name = ?", "Foo")
	DB.Last(&item, "name = 'Foo'")			// -> DB.Last(&item, "name = ?", "Foo")
	DB.Take(&item, "name = 'Foo'")			// -> DB.Take(&item, "name = ?", "Foo")
```

## Contributing

[You can help to deliver a better GORM, check out things you can do](https://gorm.io/contribute.html)

## License

© Jinzhu, 2013~time.Now

Released under the [MIT License](https://github.com/jinzhu/gorm/blob/master/License)
