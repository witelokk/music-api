package com.witelokk.tables

import org.jetbrains.exposed.sql.Table
import org.jetbrains.exposed.sql.jodatime.datetime

object PlaylistsSongs: Table("playlists_songs") {
    val playlistId = uuid("playlist_id").references(Playlists.id)
    val songsId = uuid("songs_id").references(Songs.id)
    val addedAt = datetime("added_at")
}