package com.witelokk.music.tables

import org.jetbrains.exposed.sql.Table
import org.jetbrains.exposed.sql.jodatime.datetime

object Favorites : Table("favorites") {
    val userId = uuid("user_id").references(Users.id)
    val songId = uuid("song_id").references(Songs.id)
    val addedAt = datetime("added_at")

    override val primaryKey = PrimaryKey(userId, songId)
}