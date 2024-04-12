package main

import (
	"fmt"

	raylb "github.com/gen2brain/raylib-go/raylib"
)

// The raylib window
const WINDOWWIDTH int32 = 800
const WINDOWHEIGHT int32 = 800
const INFINITY float32 = 1_000_000_000

type point struct {
	X int32
	Y int32
}

type GridWorld struct {
	Width  int
	Height int

	walls []bool // array of boolean for the walls

	start point // Path start/end points
	end   point

	// map(set) with unvisted point
	unvisited map[int32]*point
	distances []float32

	current point

	path []point
}

// function that return all the keys for a given map
// Uses Generics
// the ~ indicates derived map
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}

func (gridworld *GridWorld) wall_at(x int32, y int32) bool {
	return gridworld.walls[y*int32(gridworld.Width)+x]

}

func (gridworld *GridWorld) set_wall_at(x int32, y int32, value bool) {
	gridworld.walls[y*int32(gridworld.Width)+x] = value
}

type Rectangle raylb.Rectangle

func center_rect(rect Rectangle, relative_width float32, relative_height float32) Rectangle {

	return Rectangle{
		X:      rect.X + rect.Width*(1-relative_width)/2,
		Y:      rect.Y + rect.Height*(1-relative_height)/2,
		Width:  rect.Width * relative_width,
		Height: rect.Height * relative_height,
	}
}

// MARK: draw_grid
func draw_grid(location Rectangle, world *GridWorld) {

	var cell_width float32 = location.Width / float32(world.Width)
	var cell_height float32 = location.Height / float32(world.Height)

	var mousePosition raylb.Vector2 = raylb.GetMousePosition()

	for i := 0; i < world.Height; i++ {
		for j := 0; j < world.Width; j++ {

			cell := Rectangle{
				X:      location.X,
				Y:      location.Y,
				Width:  cell_width,
				Height: cell_height,
			}

			// Update the cell location
			cell.X += float32(j) * cell.Width
			cell.Y += float32(i) * cell.Height

			// Default color white
			color := raylb.RayWhite
			if world.wall_at(int32(j), int32(i)) {
				color = raylb.DarkPurple // set the color for the walls

			}

			var point point = point{X: int32(j), Y: int32(i)}
			if !point_is_unvisited(world, point) {
				color = raylb.LightGray

			}

			// ## Drawing the start/end points of the path
			if j == int(world.start.X) && i == int(world.start.Y) {
				color = raylb.Green

			}

			if j == int(world.end.X) && i == int(world.end.Y) {
				color = raylb.Red

			}
			// ##

			// Drawing the rectangle for the grid
			raylb.DrawRectangleRec(raylb.Rectangle(cell), color)
			raylb.DrawRectangleLinesEx(raylb.Rectangle(cell), 1, raylb.Beige)

			if !point_is_unvisited(world, point) {
				text := fmt.Sprintf("%v", distance_at(world, point))
				raylb.DrawText(text, int32(cell.X+5), int32(cell.Y+5), 25, raylb.White)
			}

			// Event handler
			// Drawing walls
			if raylb.CheckCollisionPointRec(mousePosition, raylb.Rectangle(cell)) {
				if raylb.IsMouseButtonPressed(raylb.MouseButtonLeft) {
					// method set_wall_at set the value
					// methode wall_at toggle the value btw true and false
					world.set_wall_at(int32(j), int32(i), !world.wall_at(int32(j), int32(i)))

				}

				if raylb.IsKeyReleased(raylb.KeyOne) {
					world.start = point
					reset_world(world)
				}
				if raylb.IsKeyReleased(raylb.KeyTwo) {
					world.end = point
					reset_world(world)
				}
			}

			for _, point := range world.path {
				cell := Rectangle{
					X:      location.X,
					Y:      location.Y,
					Width:  cell_width,
					Height: cell_height,
				}
				cell.X += float32(point.X) * cell_width
				cell.Y += float32(point.Y) * cell_height
				raylb.DrawRectangleLinesEx(raylb.Rectangle(cell), 5, raylb.DarkGreen)

				text := fmt.Sprintf("%v", distance_at(world, point))
				raylb.DrawText(text, int32(cell.X+5), int32(cell.Y+5), 25, raylb.Black)
			}
		}
	}

}

// MARK: Dijkstra
func step_dijkstra(world *GridWorld) {

	if !point_is_unvisited(world, world.end) {
		return
	}

	// Loop over the neighbours current node
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 { // current node
				continue
			}
			if dx != 0 && dy != 0 { // diagonal to the current node
				continue
			}
			neigbour := point{
				X: world.current.X + int32(dx),
				Y: world.current.Y + int32(dy),
			}

			if !point_in_bounds(world, neigbour) {
				continue
			}
			if !point_is_unvisited(world, neigbour) {
				continue
			}
			if world.wall_at(neigbour.X, neigbour.Y) {
				continue
			}

			var dist_current_to_neighbour float32 = 1
			distance := distance_at(world, world.current) + dist_current_to_neighbour

			if distance < distance_at(world, neigbour) {
				set_distance_at(world, neigbour, distance)
			}

		}
	}

	// ## Remove current point from unvisited
	// calculate the index
	idx := world.current.Y*int32(world.Width) + world.current.X
	_, ok := (world.unvisited[idx]) // ok = true if key exists

	if ok { // only deletes current point if it is in the unvisited map
		delete(world.unvisited, idx)
	}
	// ##

	var min_index int32 = -1

	var min_distance float32 = INFINITY

	// loop to find the minimum distance among the unvisited nodes.
	for key, point := range world.unvisited {
		dist := distance_at(world, *point)
		if dist < min_distance {
			min_distance = dist
			min_index = key
		}
	}

	if min_index != -1 {
		world.current = *world.unvisited[min_index]
	}
}

// get distance
func distance_at(world *GridWorld, point point) float32 {
	return world.distances[point.Y*int32(world.Width)+point.X]
}

// set distance
func set_distance_at(world *GridWorld, point point, value float32) {
	world.distances[point.Y*int32(world.Width)+point.X] = value
}

func point_in_bounds(world *GridWorld, point point) bool {
	return point.X >= 0 && point.X < int32(world.Width) && point.Y >= 0 && point.Y < int32(world.Height)
}

func point_is_unvisited(world *GridWorld, point point) bool {
	if len(world.unvisited) == 0 {
		return false
	}

	key := point.Y*int32(world.Width) + point.X
	_, ok := (world.unvisited[key])

	return ok

}

func reset_world(world *GridWorld) {
	for i := 0; i < world.Width*world.Height; i++ {
		world.distances[i] = INFINITY
	}

	world.current = world.start
	set_distance_at(world, world.start, 0)

	for i := 0; i < world.Height; i++ {
		for j := 0; j < world.Width; j++ {
			world.unvisited[int32(i*world.Width+j)] = &point{X: int32(j), Y: int32(i)}
		}
	}

	world.path = make([]point, 0)

}

// MARK: Reconstruct_path
func reconstruct_path(world *GridWorld) {

	if point_is_unvisited(world, world.end) {
		return
	}
	world.current = world.end
	world.path = make([]point, 0)
	world.path = append(world.path, world.current)

	for world.current.X != world.start.X || world.current.Y != world.start.Y {
		next := world.current
		min_distance := INFINITY
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}
				if dx != 0 && dy != 0 {
					continue
				}
				neigbour := point{
					X: world.current.X + int32(dx),
					Y: world.current.Y + int32(dy),
				}
				if !point_in_bounds(world, neigbour) {
					continue
				}
				if world.wall_at(neigbour.X, neigbour.Y) {
					continue
				}

				dist := distance_at(world, neigbour)
				if dist < min_distance {
					min_distance = dist
					next = neigbour
				}
			}
		}
		world.current = next
		world.path = append(world.path, world.current)

	}

}

// MARK: main
func main() {

	// Initialze the Raylib window
	raylb.InitWindow(WINDOWWIDTH, WINDOWHEIGHT, "PathFinder")
	defer raylb.CloseWindow()

	window := Rectangle{
		X:      0,
		Y:      0,
		Width:  float32(WINDOWWIDTH),
		Height: float32(WINDOWHEIGHT),
	}

	// But a subwindow within the main window
	grid_rect := center_rect(window, 0.8, 0.8)

	// Create the world
	size := 20
	walls := make([]bool, size*size)

	startpoint := point{X: 0, Y: 0}

	world := GridWorld{
		Width:  size,
		Height: size,
		walls:  walls,
		// Path start/end point
		start:     startpoint,
		end:       point{X: int32(size - 1), Y: int32(size - 1)},
		unvisited: make(map[int32]*point),

		distances: make([]float32, size*size),
		current:   startpoint,
		path:      make([]point, 0),
	}
	reset_world(&world)
	// Initialize DijKstra data
	// all points are set in the map
	// for i := 0; i < world.Height; i++ {
	// 	for j := 0; j < world.Width; j++ {
	// 		world.unvisited[int32(i*size+j)] = &point{X: int32(j), Y: int32(i)}
	// 	}
	// }
	// All distances are set to infinity except initial node
	// set_distance_at(&world, world.start, 0)
	// for i := 1; i < size*size; i++ {
	// 	world.distances[i] = INFINITY
	// }

	// Set the FPS target
	raylb.SetTargetFPS(60)

	// While loop until the window is closed
	for !raylb.WindowShouldClose() { // while loop, stop on closing window
		raylb.BeginDrawing()

		raylb.ClearBackground(raylb.DarkGray)

		// Calls the draw_grid function
		draw_grid(grid_rect, &world)

		// Do one step of Dijsktra
		// Press s key
		if raylb.IsKeyReleased(raylb.KeyS) || raylb.IsKeyDown(raylb.KeyF) {

			step_dijkstra(&world)
		}
		if raylb.IsKeyReleased(raylb.KeyR) {
			reset_world(&world)
		}

		if raylb.IsKeyReleased(raylb.KeyP) {
			reconstruct_path(&world)
		}

		raylb.DrawFPS(5, 5)

		raylb.EndDrawing()
	}
}
