import React from 'react';

import { type Artist } from './ArtistCard';

interface ArtistDetailsModalProps {
  artist: Artist | null;
  isOpen: boolean;
  onClose: () => void;
  onPlaySong?: (songName: string, artistName: string) => void;
}

export const ArtistDetailsModal: React.FC<ArtistDetailsModalProps> = ({
  artist,
  isOpen,
  onClose,
  onPlaySong,
}) => {
  if (!isOpen || !artist) {
    return null;
  }

  return (
    <div className="modal-overlay animate-fade-in">
      <div className="modal-content animate-slide-up">
        <button
          type="button"
          className="absolute top-4 right-4 text-gray-400 hover:text-white text-3xl"
          onClick={onClose}
        >
          &times;
        </button>

        <div className="text-center">
          {artist.image ? (
            <img
              src={artist.image}
              alt={artist.name}
              className="w-40 h-40 rounded-full mx-auto mb-6 border-4 border-blue-500"
            />
          ) : (
            <div className="w-40 h-40 rounded-full bg-gradient-to-br from-blue-600 to-purple-600 flex items-center justify-center text-6xl font-bold mx-auto mb-6">
              {artist.name.charAt(0)}
            </div>
          )}

          <h2 className="text-3xl font-bold mb-4">{artist.name}</h2>
          <div className="inline-block px-4 py-2 bg-blue-900/30 rounded-full mb-6">
            <span className="font-medium">{artist.genre}</span>
          </div>
          <p className="text-gray-300 mb-8 max-w-2xl mx-auto">
            {artist.description || 'Описание отсутствует'}
          </p>

          <div className="mt-8">
            <h3 className="text-2xl font-bold mb-6 text-left">Популярные треки</h3>
            <div className="space-y-3">
              {artist.songs.map((song, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between p-4 bg-gray-800/50 rounded-lg"
                >
                  <div>
                    <h4 className="font-semibold">{song.name}</h4>
                    <p className="text-gray-400 text-sm">Длительность: {song.duration}</p>
                  </div>
                  <button
                    type="button"
                    className="btn-primary"
                    onClick={() => onPlaySong?.(song.name, artist.name)}
                  >
                    ▶ Слушать
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
