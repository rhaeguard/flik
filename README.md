# flik

Gameplay so far (_screen recording has some issues_)

[https://github.com/user-attachments/assets/ce13eaa1-e275-4290-af31-5df186c0a9ab](https://github.com/user-attachments/assets/f9af3d36-0942-4fcc-a82e-504549b907ea)

### Idea

When me and my brother were little, we used to come up with ideas for different games with what we had available to us. One of those was a game where the goal was to hit and knock off all caps of the opponent. Those caps were all Coca-Cola, Fanta, Sprite caps which all had their own colors. This game was played on a chessboard.

#### Rules

- The playing field is the same as the chessboard (for now) and split into two sections, one for each player
- Each side gets 6 pieces and they can spread them as they want on their side
- A player tries to hit the opponent's cap by flicking their own cap, and the goal is to knock it off the playing field.
- You can only hit your own cap
- Game ends when only one player's caps remain on the board

### Running and Building

On Windows:

```sh
# from the root of the project
go run .
```

```sh
# from the root of the project
go build -o bin\ -ldflags "-H=windowsgui" .
# an exe should be generated in the bin directory 
```

#### References

- [vobarian - 2D collision physics article](https://www.vobarian.com/collisions/2dcollisions2.pdf)
- [pikuma - 2D collision video](https://www.youtube.com/watch?v=1L2g4ZqmFLQ)
- [elastic/inelastic collisions](https://www.youtube.com/watch?v=9YRgHikdcqs)
