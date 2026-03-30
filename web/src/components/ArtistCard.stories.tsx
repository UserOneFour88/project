import type { Meta, StoryObj } from '@storybook/react-webpack5';
import { fn } from 'storybook/test';

import { ArtistCard } from './ArtistCard';

const meta = {
  title: 'Music/ArtistCard',
  component: ArtistCard,
  tags: ['autodocs'],
  parameters: {
    layout: 'centered',
  },
  args: {
    onClick: fn(),
  },
} satisfies Meta<typeof ArtistCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Ghost: Story = {
  args: {
    artist: {
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
    },
  },
};

export const WithoutImage: Story = {
  args: {
    artist: {
      id: 2,
      name: 'Katagiri',
      genre: 'Электронная музыка',
      description: 'Электронный продюсер',
      image: null,
      songs: [
        { name: 'Tachypsychia', duration: '3:36' },
        { name: 'Synthetic heartbeat', duration: '3:12' },
      ],
    },
  },
};
