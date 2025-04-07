# flik

Gameplay so far (_screen recording has some issues_)

https://github.com/user-attachments/assets/62f37ef5-ee2e-40e0-b870-7527229fe023

### Idea

When me and my brother were little, we used to come up with ideas for different games with what we had available to us. One of those was a game where the goal was to hit and knock off all bottle caps of the opponent from the board. Those bottle caps were all Coca-Cola, Fanta, Sprite caps which all had their own colors. We used to play it on a chessboard.

#### Rules

- Each side gets 6 pieces that are spread randomly into 6 of possible 12 positions
- A player tries to hit the opponent's cap by launching their own cap towards the opponent, and the goal is to knock caps off the playing field.
- You can only launch your own cap
- Game ends when only one player's caps remain on the board
- Caps have life points and both hitting and getting hit takes life points

### Running and Building

On Windows:

There's already a Windows executable provided in the [bin](./bin) folder. Alternatively you can build and run from the source.

```sh
# from the root of the project
go run .
```

```sh
# from the root of the project
go build -o bin\ -ldflags "-H=windowsgui -s -w" -tags release .
# an exe should be generated in the bin directory 
```

## Credits

- Background music (_Sketchbook 2024-11-29_) by Abstraction ([website](https://abstractionmusic.com/), [music-loop-bundle](https://tallbeard.itch.io/music-loop-bundle)) 

#### References

- [vobarian - 2D collision physics article](https://www.vobarian.com/collisions/2dcollisions2.pdf)
- [pikuma - 2D collision video](https://www.youtube.com/watch?v=1L2g4ZqmFLQ)
- [elastic/inelastic collisions](https://www.youtube.com/watch?v=9YRgHikdcqs)

