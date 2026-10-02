/**
 * @file Tournament.cpp
 * @brief Реализация базового класса турнира.
 */

#include "Tournament.h"

Tournament::Tournament(
    const std::string& name,
    int maxParticipants
)
    : name(name),
      maxParticipants(maxParticipants),
      participantsCount(0),
      published(false)
{
}

void Tournament::publish()
{
    published = true;
}

bool Tournament::addParticipant()
{
    if (participantsCount >= maxParticipants)
    {
        return false;
    }

    ++participantsCount;
    return true;
}

int Tournament::getParticipantsCount() const
{
    return participantsCount;
}