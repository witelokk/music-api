package models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class Artist(
    val id: String,
    val name: String,
    @SerialName("avatar_url") val avatarUrl: String?,
    @SerialName("cover_url") val coverUrl: String?,
    val followers: Int,
    val following: Boolean,
)