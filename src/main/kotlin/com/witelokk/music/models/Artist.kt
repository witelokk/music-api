package com.witelokk.music.models

import com.witelokk.music.UUIDSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.util.*

@Serializable
data class Artist(
    @Serializable(UUIDSerializer::class) val id: UUID,
    val name: String,
    @SerialName("avatar_url") val avatarUrl: String?,
    @SerialName("cover_url") val coverUrl: String?,
    val followers: Int,
    val following: Boolean,
)