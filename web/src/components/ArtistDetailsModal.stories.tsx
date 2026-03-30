import type { Meta, StoryObj } from '@storybook/react-webpack5';
import React, { useState } from 'react';
import { fn } from 'storybook/test';

import { type Artist,ArtistCard } from './ArtistCard';
import { ArtistDetailsModal } from './ArtistDetailsModal';

const demoArtist: Artist = {
  id: 1,
  name: 'Ghost',
  genre: 'Рок',
  description: 'Шведская рок-группа из Линчёпинга',
  image: 'images/ghost.jpg',
  songs: [
    { name: 'Mary on a cross', duration: '4:56' },
    { name: "Hunter's Moon", duration: '4:03' },
    { name: 'He is', duration: '3:05' },
  ],
};

const meta = {
  title: 'Music/ArtistDetailsModal',
  component: ArtistDetailsModal,
  tags: ['autodocs'],
  parameters: {
    layout: 'fullscreen',
  },
  args: {
    artist: null,
    isOpen: false,
    onClose: fn(),
    onPlaySong: fn(),
  },
} satisfies Meta<typeof ArtistDetailsModal>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Open: Story = {
  args: {
    isOpen: true,
    artist: demoArtist,
  },
};

const InteractivePreview: React.FC = () => {
  const [isOpen, setIsOpen] = useState<boolean>(false);
  const [selectedArtist, setSelectedArtist] = useState<Artist | null>(null);

  const handleOpen = (artist: Artist): void => {
    setSelectedArtist(artist);
    setIsOpen(true);
  };

  const handleClose = (): void => {
    setIsOpen(false);
    setSelectedArtist(null);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 to-black text-white p-8">
      <div className="max-w-xs">
        <ArtistCard artist={demoArtist} onClick={handleOpen} />
      </div>
      <ArtistDetailsModal
        artist={selectedArtist}
        isOpen={isOpen}
        onClose={handleClose}
        onPlaySong={fn()}
      />
    </div>
  );
};

export const ClickCardToOpen: Story = {
  args: {
    artist: null,
    isOpen: false,
  },
  render: () => <InteractivePreview />,
};
