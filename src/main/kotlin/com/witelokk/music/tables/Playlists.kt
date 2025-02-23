package com.witelokk.music.tables

import org.jetbrains.exposed.sql.Table
import org.jetbrains.exposed.sql.jodatime.datetime

object Playlists: Table("playlists") {
    val id = uuid("id").uniqueIndex()
    val userId = uuid("user_id").references(Users.id)
    val name = varchar("name", 255)
    val createdAt = datetime("created_at")
}
