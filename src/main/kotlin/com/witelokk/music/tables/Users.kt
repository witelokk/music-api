package com.witelokk.music.tables

import org.jetbrains.exposed.sql.Table
import org.jetbrains.exposed.sql.jodatime.datetime

object Users : Table("users") {
    val id = uuid("id").uniqueIndex()
    val name = varchar("name", 255)
    val email = text("email").uniqueIndex()
    val createdAt = datetime("created_at")
}
