package com.witelokk.tables

import org.jetbrains.exposed.sql.Table
import org.jetbrains.exposed.sql.jodatime.datetime

object Users : Table("users") {
    val id = uuid("id")
    val name = varchar("name", 255).uniqueIndex()
    val email = text("email")
    val createdAt = datetime("created_at")
}
