import React, { useState } from 'react';

import { type Artist, ArtistCard } from './components/ArtistCard';
import { ArtistDetailsModal } from './components/ArtistDetailsModal';

interface Track {
  name: string;
  duration: string;
  artist: string;
  genre: string;
}

// Импорт изображений с типами

const artistsData: Artist[] = [
  {
    id: 1,
    name: "Ghost",
    genre: "Рок",
    description: "Шведская рок-группа из Линчёпинга...",
    image: "images/ghost.jpg",
    songs: [
      { name: "Mary on a cross", duration: "4:56" },
      { name: "Hunter's Moon", duration: "4:03" },
      { name: "He is", duration: "3:05" },
    ],
  },
  {
    id: 2,
    name: "Katagiri",
    genre: "Электронная музыка",
    description: "Электронный продюсер",
    image: null,
    songs: [
      { name: "Tachypsychia", duration: "3:36" },
      { name: "Synthetic heartbeat", duration: "3:12" },
      { name: "Neon drift", duration: "4:01" },
    ],
  },
  {
    id: 3,
    name: "Aether Realm",
    genre: "Метал",
    description: "Американская фолк-метал группа",
    image: "images/aether-realm.jpg",
    songs: [
      { name: "The sun The moon The star", duration: "17:28" },
      { name: "The Magician", duration: "4:22" },
      { name: "Guardian", duration: "5:01" },
    ],
  },
  {
    id: 4,
    name: "Breaking Benjamin",
    genre: "Рок",
    description: "Американская рок-группа",
    image: "images/breaking-benjamin.jpg",
    songs: [
      { name: "Dance With The Devil", duration: "4:45" },
      { name: "The Diary of Jane", duration: "3:20" },
      { name: "Breath", duration: "3:38" },
    ],
  },
  {
    id: 5,
    name: "Polyphia",
    genre: "Прочее",
    description: "Инструментальный прогрессив-метал, США",
    image: "images/polyphia.jpg",
    songs: [
      { name: "Playing God", duration: "3:23" },
      { name: "G.O.A.T.", duration: "4:09" },
      { name: "Ego Death", duration: "3:45" },
    ],
  },
  {
    id: 6,
    name: "System of a Down",
    genre: "Метал",
    description: "Армяно-американская альтернативная метал-группа",
    image: "images/system-of-a-down.jpg",
    songs: [
      { name: "Chop Suey!", duration: "3:30" },
      { name: "Toxicity", duration: "3:38" },
      { name: "Aerials", duration: "3:55" },
    ],
  },
  {
    id: 7,
    name: "The Living Tombstone",
    genre: "Электронная музыка",
    description: "Электронный проект, известен по фан-трекам",
    image: "images/the-living-tombstone.jpg",
    songs: [
      { name: "Discord", duration: "4:10" },
      { name: "Five Nights at Freddy's", duration: "4:02" },
      { name: "My Ordinary Life", duration: "3:48" },
    ],
  },
  {
    id: 8,
    name: "Theory of a Deadman",
    genre: "Рок",
    description: "Канадская рок-группа",
    image: "images/theory-of-a-deadman.jpg",
    songs: [
      { name: "Bad Girlfriend", duration: "3:25" },
      { name: "Rx (Medicate)", duration: "3:02" },
      { name: "Not Meant to Apologize", duration: "3:44" },
    ],
  },
];

const App: React.FC = () => {
  const [currentFilter, setCurrentFilter] = useState<string>('all');
  const [selectedArtist, setSelectedArtist] = useState<Artist | null>(null);
  const [isModalOpen, setIsModalOpen] = useState<boolean>(false);

  // Получение всех треков
  const getAllTracks = (): Track[] => {
    const tracks: Track[] = [];
    artistsData.forEach(artist => {
      artist.songs.forEach(song => {
        tracks.push({
          name: song.name,
          duration: song.duration,
          artist: artist.name,
          genre: artist.genre
        });
      });
    });
    return tracks;
  };

  const allTracks = getAllTracks();
  const filteredTracks = currentFilter === 'all' 
    ? allTracks 
    : allTracks.filter(track => track.genre === currentFilter);

  const handleFilterClick = (genre: string): void => {
    setCurrentFilter(genre);
  };

  const handleArtistClick = (artist: Artist): void => {
    setSelectedArtist(artist);
    setIsModalOpen(true);
  };

  const handlePlaySong = (songName: string, artistName: string): void => {
    alert(`🎵 Воспроизведение: ${songName} - ${artistName}`);
  };

  const handleCloseModal = (): void => {
    setIsModalOpen(false);
    setSelectedArtist(null);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 to-black text-white">
      <header className="py-12 text-center">
        <h1 className="text-5xl font-bold mb-4 bg-gradient-to-r from-blue-400 to-purple-500 bg-clip-text text-transparent">
          Музыкальная<br />коллекция
        </h1>
        <p className="text-gray-400 text-lg">Сайт для дисциплины web-программирование</p>
      </header>

      <main className="container mx-auto px-4">
        {/* Фильтры */}
        <div className="flex flex-wrap gap-3 justify-center mb-10">
          {['all', 'Рок', 'Электронная музыка', 'Метал', 'Прочее'].map(genre => (
            <button
              key={genre}
              className={`px-6 py-3 rounded-full font-medium transition-all duration-300 ${
                currentFilter === genre
                  ? 'bg-gradient-to-r from-blue-600 to-blue-700 text-white shadow-lg'
                  : 'bg-gray-800 text-gray-300 hover:bg-gray-700'
              }`}
              onClick={() => handleFilterClick(genre)}
            >
              {genre === 'all' ? 'Все жанры' : genre}
            </button>
          ))}
        </div>

        {/* Секция треков */}
        <div className="card mb-12">
          <h2 className="text-2xl font-bold mb-6 text-center">
            {currentFilter === 'all' ? 'Все треки' : `Треки в жанре: ${currentFilter}`}
          </h2>
          <div className="space-y-4">
            {filteredTracks.length === 0 ? (
              <p className="text-center text-gray-500 italic">Треки не найдены</p>
            ) : (
              filteredTracks.map((track, index) => (
                <div 
                  key={index} 
                  className="flex items-center justify-between p-4 bg-gray-800/50 rounded-lg hover:bg-gray-800 transition-colors group"
                >
                  <div>
                    <h3 className="font-semibold text-lg">{track.name}</h3>
                    <p className="text-gray-400">{track.artist}</p>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className="text-gray-500">{track.duration}</span>
                    <button 
                      className="btn-primary flex items-center gap-2"
                      onClick={() => handlePlaySong(track.name, track.artist)}
                    >
                      <span>▶</span>
                      <span>Слушать</span>
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Секция исполнителей */}
        <h2 className="text-3xl font-bold mb-8 text-center">
          Исполнители <span className="text-blue-400">({artistsData.length})</span>
        </h2>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-12">
          {artistsData.map(artist => (
            <ArtistCard key={artist.id} artist={artist} onClick={handleArtistClick} />
          ))}
        </div>
      </main>

      <ArtistDetailsModal
        artist={selectedArtist}
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        onPlaySong={handlePlaySong}
      />
    </div>
  );
};

export default App;