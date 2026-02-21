"use client";
import { Center, Container, Table, Text } from '@chakra-ui/react';

export default function Home() {
  const playersAndPoints = [
    { id: 1, name: "Ethan", points: 25},
    { id: 4, name: "Ian", points: 23 },
    { id: 2, name: "Gordie", points: 23 },
    { id: 5, name: "Cam", points: 22},
    { id: 6, name: "Mike", points: 21 },
    { id: 3, name: "Tay", points: 20},
    { id: 7, name: "Cho", points: 15}
  ]
  
  return (
    <Container>

      <Text textAlign="center" fontSize="5xl">CFB Manhole</Text>
      <Container fluid maxW={"lg"}>
        <Table.Root size='sm' variant="outline">
          <Table.Header>
            <Table.Row>
              <Table.ColumnHeader>Names</Table.ColumnHeader>
              <Table.ColumnHeader>Points</Table.ColumnHeader>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {playersAndPoints.map((player) => (
              <Table.Row key={ player.id }>
                <Table.Cell>{player.name}</Table.Cell>
                <Table.Cell>{player.points}</Table.Cell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table.Root>
      </Container>
    </Container>
  );
}
