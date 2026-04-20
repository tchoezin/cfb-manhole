"use client";
import { Container, Spinner, Table, Text, VStack } from '@chakra-ui/react';
import { useEffect, useMemo, useState } from 'react';

import { ApiError, cfbApi } from '@/lib/api';

type LeaderboardRow = {
  id: string;
  name: string;
  points: number;
  division: string;
};

export default function Home() {
  const [rows, setRows] = useState<LeaderboardRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;

    const loadLeaderboard = async () => {
      try {
        setLoading(true);
        setError(null);

        const response = await cfbApi.getLeaderboard('default');
        const flattenedRows: LeaderboardRow[] = [];

        for (const [divisionName, divisionRows] of Object.entries(response.leaderboard)) {
          for (const row of divisionRows) {
            flattenedRows.push({
              id: `${divisionName}:${row.player}`,
              name: row.player,
              points: row.score,
              division: divisionName,
            });
          }
        }

        flattenedRows.sort((a, b) => b.points - a.points || a.name.localeCompare(b.name));

        if (isMounted) {
          setRows(flattenedRows);
        }
      } catch (err) {
        if (!isMounted) {
          return;
        }

        if (err instanceof ApiError) {
          setError(`API Error (${err.status})`);
        } else {
          setError('Failed to load leaderboard');
        }
      } finally {
        if (isMounted) {
          setLoading(false);
        }
      }
    };

    loadLeaderboard();

    return () => {
      isMounted = false;
    };
  }, []);

  const hasRows = useMemo(() => rows.length > 0, [rows]);

  return (
    <Container>
      <Text textAlign="center" fontSize="5xl">CFB Manhole</Text>

      {loading ? (
        <VStack py={10}>
          <Spinner size="lg" />
          <Text>Loading leaderboard...</Text>
        </VStack>
      ) : null}

      {error ? (
        <Text color="red.500" textAlign="center" py={4}>
          {error}
        </Text>
      ) : null}

      <Container fluid maxW={"lg"}>
        <Table.Root size='sm' variant="outline">
          <Table.Header>
            <Table.Row>
              <Table.ColumnHeader>Names</Table.ColumnHeader>
              <Table.ColumnHeader>Division</Table.ColumnHeader>
              <Table.ColumnHeader>Points</Table.ColumnHeader>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {hasRows ? rows.map((player) => (
              <Table.Row key={player.id}>
                <Table.Cell>{player.name}</Table.Cell>
                <Table.Cell>{player.division}</Table.Cell>
                <Table.Cell>{player.points}</Table.Cell>
              </Table.Row>
            )) : !loading ? (
              <Table.Row>
                <Table.Cell>No players yet</Table.Cell>
                <Table.Cell>-</Table.Cell>
                <Table.Cell>0</Table.Cell>
              </Table.Row>
            ) : null}
          </Table.Body>
        </Table.Root>
      </Container>
    </Container>
  );
}
