"use client";
import { FetchRankings, RankingEntry } from '@/lib/api';
import { Container, Table, Text } from '@chakra-ui/react';
import { useEffect, useState } from "react";

export default function Home() {
 const [rankings, setRankings] = useState<RankingEntry[]>([]);
 const [loading, setLoading] = useState(true);
 const [error, setError] = useState<string | null>(null);

 useEffect(() => {
  let cancelled = false;

  async function loadRankings() {
    try {
      setLoading(true);
      setError(null);

      const rankings = await FetchRankings()

      if (!cancelled) {
        setRankings(rankings);
      }
    } catch (err) {
      if (!cancelled) {
        setError(err instanceof Error ? err.message : "Unknown error");
      }
    } finally {
      if (!cancelled) {
        setLoading(false);
      }
    }
  }

  loadRankings();

  return () => {
    cancelled = true;
  };
 }, []);

 return (
  <Container>
    <Text textAlign="center" fontSize="5xl" mb={6}>
      CFB Manhole
    </Text>

    {loading && (
      <Text textAlign="center" mb={4}>
        Loading rankings...
      </Text>
    )}

    {error && (
      <Text textAlign="center" color="red.500" mb={4}>
        Error: {error}
      </Text>
    )}

    <Container maxW="lg">
      <Table.Root size="sm" variant="outline">
        <Table.Header>
          <Table.Row>
            <Table.ColumnHeader>Rank</Table.ColumnHeader>
            <Table.ColumnHeader>Player</Table.ColumnHeader>
            <Table.ColumnHeader>Points</Table.ColumnHeader>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {rankings.map((entry) => (
            <Table.Row key={entry.player + "-" + entry.rank}>
              <Table.Cell>{entry.rank}</Table.Cell>
              <Table.Cell>{entry.player}</Table.Cell>
              <Table.Cell>{entry.score}</Table.Cell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table.Root>
    </Container>
  </Container>
 )
}
