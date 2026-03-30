module.exports = {
preset: 'ts-jest',
testEnvironment: 'jsdom',
setupFilesAfterEnv: ['@testing-library/jest-dom/extend-expect'],
moduleNameMapper: {
'^@/(.*)$': '<rootDir>/src/$1',
},
testMatch: ['<rootDir>/src/**/*.test.{ts,tsx}'],
};