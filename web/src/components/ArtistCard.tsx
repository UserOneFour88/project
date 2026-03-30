import React from 'react';

export interface Song {
  name: string;
  duration: string;
}

export interface Artist {
  id: number;
  name: string;
  genre: string;
  description: string;
  image: string | null;
  songs: Song[];
}

interface ArtistCardProps {
  artist: Artist;
  onClick?: (artist: Artist) => void;
}

export const ArtistCard: React.FC<ArtistCardProps> = ({ artist, onClick }) => {
  return (
    <button
      type="button"
      className="card hover:scale-105 transition-transform cursor-pointer w-full"
      onClick={() => onClick?.(artist)}
    >
      <div className="flex flex-col items-center text-center">
        {artist.image ? (
          <img
            src={artist.image}
            alt={artist.name}
            className="w-32 h-32 rounded-full object-cover mb-4 border-4 border-blue-500/30"
          />
        ) : (
          <div className="w-32 h-32 rounded-full bg-gradient-to-br from-blue-600 to-purple-600 flex items-center justify-center text-4xl font-bold mb-4">
            {artist.name.charAt(0)}
          </div>
        )}
        <h3 className="text-xl font-bold mb-2">{artist.name}</h3>
        <span className="px-3 py-1 bg-gray-800 rounded-full text-sm">{artist.genre}</span>
      </div>
    </button>
  );
};
