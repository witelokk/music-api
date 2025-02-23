package com.witelokk.music.tables

import org.jetbrains.exposed.sql.Table

object Artists: Table("artists") {
    val id = uuid("id").uniqueIndex()
    val name = varchar("name", 255)
    val avatarUrl = text("avatar_url").nullable()
    val coverUrl = text("cover_url").nullable()
}
