package agent

import (
	"vec"
	"fmt"
	"math"
	"encoding/json"
	"sync"
	"strings"
)

type Agent struct {
	sync.RWMutex
	odin *Odin
	Type string `json:"type"`
	Id int64 `json:"id"`

	Pos vec.Vec3 `json:"pos"`
	Target vec.Vec3 `json:"target"`
	forward vec.Vec3
	roll float64
	Euler vec.Euler `json:"euler"`

	velocity vec.Vec3
	dist_delta vec.Vec3
	mass float64
	radius float64

	alive bool
	age int64
	lifespan int64

	rocket Rocket
}

func (agent *Agent) findClosestPlant(ents []Entity) Entity{
	var closest_dist float64 = math.MaxFloat64
	var closest Entity = nil

	for _, ent := range ents {
		if agent != ent {
			if strings.Compare(ent.GetType(), "plant") == 0 {
				dist := vec.Vec3DistanceBetween(agent.GetPos(), ent.GetPos())
				if dist < closest_dist {
					dist = closest_dist
					closest = ent
				}
			}
		}
	}
	return closest
}

func (agent *Agent) calculateMovement(time_delta float64) {
	var forward_norm = vec.Vec3Normal(agent.forward)

	var up = vec.Vec3{0.0, 1.0, 0.0}
	var right = vec.Vec3Cross(forward_norm, up)
	up = vec.Vec3Cross(forward_norm, right)

	agent.turn(time_delta)
	agent.thrustForward(time_delta)
	//agent.stabilize(time_delta)

	agent.dist_delta = vec.Vec3Scale(agent.velocity, time_delta)

	if vec.Vec3DistanceBetween(agent.Target, agent.Pos) < 2000 {
		agent.Target = vec.Vec3Add(agent.Pos, vec.Vec3Random(95000))
		//agent.Target = vec.Vec3Random(35000)
	}
}

func (agent *Agent) turn(time_delta float64) {
	var course = vec.Vec3Sub(agent.Target, agent.Pos)
	var course_normal = vec.Vec3Normal(course)

	var angle_diff = vec.Vec3AngleBetween(agent.forward, course_normal)
	if angle_diff > math.Pi {
		angle_diff = math.Pi
	}
	var axis = vec.Vec3Cross(agent.forward, course_normal)

	var delta_turn = time_delta
	var new_forward = vec.QuaternionRotation(agent.forward,  delta_turn, axis)
	agent.forward = new_forward
}

func (agent *Agent) thrustForward(time_delta float64) {
	impulse := agent.rocket.Thrust(vec.Vec3Normal(agent.forward), 1.0, time_delta)
	agent.applyImpulse(impulse)
}

func (agent *Agent) stabilize(time_delta float64) {
	var velocity_normal = vec.Vec3Normal(agent.velocity)
	var course = vec.Vec3Sub(agent.Target, agent.Pos)
	var course_relative = vec.Vec3Sub(course, agent.Pos)
	var course_normal = vec.Vec3Normal(course_relative)
	var force_dir = vec.QuaternionRotation(velocity_normal, math.Pi, course_normal)

	impulse := agent.rocket.Thrust(force_dir, 0.4, time_delta)
	agent.applyImpulse(impulse)
}

func (agent *Agent) applyForce(force vec.Vec3, time_delta float64) {
	agent.Lock()
	agent.applyImpulse(vec.Vec3Scale(force, time_delta))
	agent.Unlock()
}

func (agent *Agent) applyImpulse(impulse vec.Vec3) {
	agent.Lock()
	velo_delta := vec.Vec3Scale(impulse, 1 / agent.mass)
	agent.velocity = vec.Vec3Add(agent.velocity, velo_delta)
	agent.Unlock()
}

func (agent *Agent) Move(time_delta float64) {
	dist := vec.Vec3Scale(agent.velocity, time_delta)
	agent.Pos = vec.Vec3Add(agent.Pos, dist)
}

func (agent *Agent) Act(time_delta float64) {
	fmt.Println("agent acted")
}

func (agent *Agent) Alive() bool {
	return agent.alive
}

func (agent *Agent) GetPos() vec.Vec3 {
	return agent.Pos
}

func (agent *Agent) SetPos(new_pos vec.Vec3) {
	agent.Lock()
	agent.Pos = new_pos
	agent.Unlock()
}

func (agent *Agent) GetVelocity() vec.Vec3 {
	return agent.velocity
}

func (agent *Agent) AddVelocity(delta vec.Vec3) {
	agent.Lock()
	agent.velocity = vec.Vec3Add(agent.velocity, delta)
	agent.Unlock()
}

func (agent *Agent) GetMass() float64 {
	return agent.mass
}

func (agent *Agent) GetRadius() float64 {
	return agent.radius
}

func (agent *Agent) GetType() string {
	return agent.Type
}

func (agent *Agent) GetJSON() string {
	json, _ := json.Marshal(agent)
	return string(json)
}


